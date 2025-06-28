package tasks

import (
	"fmt"
	"runtime"
	"sync"

	tasks "popmart/src/backend/tasks"
	helpers "popmart/src/middleware/helpers"
	desktop "popmart/src/middleware/modules/desktop"

	"github.com/AlecAivazis/survey/v2"
)

func TasksMenu(logger *helpers.ColorizedLogger) {
	for {
		var result string
		options := []string{
			"Start Tasks",
			"Open Tasks",
			"Back",
		}

		mainPrompt := &survey.Select{
			Message: "Tasks Menu:",
			Options: options,
		}

		err := survey.AskOne(mainPrompt, &result)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed To Prompt Tasks Menu: %v", err))
			return
		}

		switch result {
		case "Start Tasks":
			groups, err := tasks.LoadTaskGroups()
			if err != nil {
				logger.Error("Failed To Load Task Groups: " + err.Error())
				continue
			}

			if len(groups) == 0 {
				logger.Warn("No Task Groups Found In tasks.csv")
				continue
			}

			var selectedGroup string
			groupPrompt := &survey.Select{
				Message: "Select Task Group To Start:",
				Options: groups,
			}

			err = survey.AskOne(groupPrompt, &selectedGroup)
			if err != nil {
				logger.Error("Prompt Cancelled Or Failed: " + err.Error())
				continue
			}

			loadedTasks, err := tasks.LoadTasks(logger, selectedGroup)
			if err != nil {
				logger.Error("Failed To Load Tasks: " + err.Error())
				continue
			}

			maxWorkers := helpers.CalculateWorkers()
			logger.Info(fmt.Sprintf("Starting %d Workers Based On %d CPU Cores", maxWorkers, runtime.NumCPU()))

			var wg sync.WaitGroup
			tasksChan := make(chan helpers.Task)

			for range make([]struct{}, maxWorkers) {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for t := range tasksChan {
						switch t.Mode {
						case "Desktop":
							desktop.PopmartDesktop(t, logger)
						case "App":
							logger.Error("App Mode Is Currently Down For Maintenance")
						default:
							logger.Warn(fmt.Sprintf("Task %s: Unsupported Task Mode Has Been Declared", t.TaskId))
						}
					}
				}()
			}

			for _, task := range loadedTasks {
				tasksChan <- task
			}
			close(tasksChan)

			wg.Wait()
		case "Open Tasks":
			err := tasks.OpenTasksCSV()
			if err != nil {
				logger.Error("Failed To Open Tasks: " + err.Error())
				continue
			}
			logger.Silly("Opened Tasks CSV In Default Editor")

		case "Back":
			return

		default:
			logger.Warn("Invalid option selected")
		}
	}
}
