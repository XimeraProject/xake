package main

import (
	"fmt"
	"gopkg.in/cheggaaa/pb.v1"
	"sync"
)

func Bake(workers int) error {
	tasks := make(chan string)

	fmt.Printf("Using %d workers", workers)

	files, err := NeedingCompilation(repository)
	// BADBAD: need to display error from compilation if it fails
	if err != nil {
		return err
	}

	bar := pb.StartNew(len(files))
	bar.SetMaxWidth(80)
	bar.ShowTimeLeft = true
	bar.Start()

	var group sync.WaitGroup
	for i := 0; i < workers; i++ {
		group.Add(1)
		go func() {
			for task := range tasks {
				Compile(repository, task)
				bar.Increment()
			}
			group.Done()
		}()
	}

	for _, file := range files {
		tasks <- file
	}
	close(tasks)

	group.Wait()
	bar.FinishPrint("The xake is made.")

	return nil
}
