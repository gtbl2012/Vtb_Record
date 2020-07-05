package main

import (
	"fmt"
	"github.com/fzxiao233/Vtb_Record/live"
	"github.com/fzxiao233/Vtb_Record/live/monitor"
	. "github.com/fzxiao233/Vtb_Record/utils"
	"github.com/orandin/lumberjackrus"
	log "github.com/sirupsen/logrus"
	"time"
)

// Can't be func init as we need the parsed config
func initLog() {
	log.Printf("Init logging!")
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{})
	level := log.InfoLevel
	if Config.LogLevel == "debug" {
		level = log.DebugLevel
	} else if Config.LogLevel == "info" {
		level = log.InfoLevel
	} else if Config.LogLevel == "warn" {
		level = log.WarnLevel
	} else if Config.LogLevel == "error" {
		level = log.ErrorLevel
	}
	log.SetLevel(level)
	hook, err := lumberjackrus.NewHook(
		&lumberjackrus.LogFile{
			Filename:   Config.LogFile,
			MaxSize:    Config.LogFileSize,
			MaxBackups: 1,
			MaxAge:     1,
			Compress:   false,
			LocalTime:  false,
		},
		level,
		&log.JSONFormatter{},
		nil,
	)

	if err != nil {
		panic(fmt.Errorf("NewHook Error: %s", err))
	}

	log.AddHook(hook)
}

func arrangeTask() {
	log.Printf("Arrange tasks...")
	status := make([][]bool, len(Config.Module))
	for i, module := range Config.Module {
		status[i] = make([]bool, len(module.Users))
		for j, _ := range status[i] {
			status[i][j] = false
		}
	}

	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(1))
		for {
			if ConfigChanged {
				allDone := true
				/*for mod_i, _ := range status {
					for _, ch := range status[mod_i] {
						if ch != false {
							allDone = false
						}
					}
				}*/
				if allDone {
					ret, err := ReloadConfig()
					if ret {
						if err == nil {
							log.Infof("Config changed! New config: %s", Config)
						} else {
							log.Warnf("Config changed but loading failed: %s", err)
						}
					}
				}
			}
			<-ticker.C
		}

	}()

	for {
		for mod_i, module := range Config.Module {
			if module.Enable {
				for user_i, usersConfig := range module.Users {
					if status[mod_i][user_i] != false {
						continue
					}
					status[mod_i][user_i] = true
					//log.Printf("%s|%s is up", module.Name, usersConfig.Name)
					go func(i, j int, mon monitor.VideoMonitor, userCon UsersConfig) {
						live.StartMonitor(mon, userCon)
						status[i][j] = false
					}(mod_i, user_i, monitor.CreateVideoMonitor(module), usersConfig)
					time.Sleep(time.Millisecond * time.Duration(10))
				}
			}
		}
		if time.Now().Minute() > 55 || time.Now().Minute() < 5 || (time.Now().Minute() > 25 && time.Now().Minute() < 35) {
			time.Sleep(time.Duration(Config.CriticalCheckSec) * time.Second)
		}
		time.Sleep(time.Duration(Config.NormalCheckSec) * time.Second)
	}
}
func main() {
	initLog()
	arrangeTask()
}
