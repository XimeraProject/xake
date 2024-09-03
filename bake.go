package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/cheggaaa/pb.v1"
	"os"
	"sync"
)

type compilationResult struct {
	name string
	succes bool
}

func Bake(workers int, compilationChecker func(repository string) ([]string, map[string][]string, error), compiler func(repository string, task string) ([]byte, error)) error {
	tasks := make(chan string)
	queue := make(chan string)

	log.Debug(fmt.Sprintf("Using %d workers", workers))

	files, dependencies, err := compilationChecker(repository)
	// BADBAD: need to display error from compilation if it fails
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	finished := make(chan compilationResult, len(files))
	finishedCount := 0
	failedCount := 0
	compilationSkippedCount := 0

	var bar *pb.ProgressBar
	if log.Level != logrus.DebugLevel {
		bar = pb.StartNew(len(files))
		//bar.SetMaxWidth(80)
		bar.ShowTimeLeft = true
		bar.Start()
	}

	// Manage a pool of workers
	var group sync.WaitGroup
	for i := 0; i < workers; i++ {
		group.Add(1)
		go func(workerId int) {
			log.Debug(fmt.Sprintf("Worker %d is running", workerId))
			for task := range tasks {
				log.Debug(fmt.Sprintf("Worker %d is compiling %s", workerId, task))
				_, err = compiler(repository, task)

				if err != nil {
					log.Error(fmt.Sprintf("Worker %d could not compile %s", workerId, task))
					finished <- compilationResult{task, false}
					failedCount++
				} else {
					log.Debug(fmt.Sprintf("Worker %d finishes with %s", workerId, task))
					finished <- compilationResult{task, true}
					finishedCount++
				}
				if log.Level != logrus.DebugLevel {
					bar.Increment()
				} else {
					log.Info(fmt.Sprintf("Finished %d/%d tasks: %d succes, %d failed, %d skipped", finishedCount + failedCount + compilationSkippedCount, len(files), finishedCount, failedCount, compilationSkippedCount))
				}
			}
			group.Done()
		}(i + 1)
	}

	// As things arrive in the queue, send them out to the various workers
	go func() {
		for file := range queue {
			log.Debug("Sending " + file + " to workers")
			tasks <- file
		}
		close(tasks)
	}()

	compiled := make(map[string]bool)
	compilationFailed := make(map[string]bool)
	compilationSkipped := make(map[string]bool)
	enqueued := make(map[string]bool)
	enqueuedCount := 0

	for {
		// add everything that can be compiled given what has been completed
		for _, file := range files {
			if enqueued[file] {
				continue
			}

			good := true
			for _, dependency := range dependencies[file] {
				if compilationFailed[dependency] || compilationSkipped[dependency] {
					log.Info("Skipping " + file + " because failure of " + dependency)
					compilationSkippedCount++
					compilationSkipped[file] = true
					// Pretend it is enqueued
					enqueued[file] = true
					enqueuedCount++
					good = false
					break
				}
			}
			if good {
				for _, dependency := range dependencies[file] {

					if !compiled[dependency] {
						good = false
						break
					}
				}
			}

			if good {
				log.Debug("Placing " + file + " into the queue")
				queue <- file
				enqueued[file] = true
				enqueuedCount++
			}
		}

		log.Debug(fmt.Sprintf("%d tasks have been placed in the queue.", enqueuedCount))
		if enqueuedCount >= len(files) {
			log.Debug("Everything is in the queue.")
			break
		}

		log.Debug("Wait for things to finish before placing more into the queue.")
		// get more finished things
		finishedTask := <-finished
		if finishedTask.succes {
			compiled[finishedTask.name] = true
		} else {
			compilationFailed[finishedTask.name] = true
		}
	}

	close(queue)
	log.Debug("Waiting for the workers to finish.")
	group.Wait()
	if failedCount > 0 {
		log.Error("Failed baking the whole xake.")
		for name, val := range compilationFailed {
			if val {
				log.Error("Failed compiling " + name)
			}
		}
		for name, val := range compilationSkipped {
			if val {
				log.Info("Skipped compilation " + name)
			}
		}
	        os.Exit(1)
        }

	if log.Level != logrus.DebugLevel {
		bar.FinishPrint("The xake is made.")
	} else {
		log.Info("The xake is made.")
	}

	return nil
}
