package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/cheggaaa/pb.v1"
	"os"
	"sync"
)

func Bake(workers int) error {
	tasks := make(chan string)
	queue := make(chan string)

	log.Debug(fmt.Sprintf("Using %d workers", workers))

	files, dependencies, err := NeedingCompilation(repository)
	// BADBAD: need to display error from compilation if it fails
	if err != nil {
		return err
	}

	finished := make(chan string, len(files))
	finishedCount := 0

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
				_, err := Compile(repository, task)

				if err != nil {
					log.Error("Could not compile " + task)
					os.Exit(1)
				} else {
					log.Debug(fmt.Sprintf("Worker %d finishes with %s", workerId, task))

					finished <- task
					finishedCount++

					if log.Level != logrus.DebugLevel {
						bar.Increment()
					} else {
						log.Info(fmt.Sprintf("Finished %d/%d tasks", finishedCount, len(files)))
					}
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
				if !compiled[dependency] {
					good = false
					break
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
		compiled[<-finished] = true
	}

	close(queue)
	log.Debug("Waiting for the workers to finish.")
	group.Wait()

	if log.Level != logrus.DebugLevel {
		bar.FinishPrint("The xake is made.")
	} else {
		log.Info("The xake is made.")
	}

	return nil
}
