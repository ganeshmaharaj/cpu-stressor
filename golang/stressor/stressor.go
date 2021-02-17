package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"
)

/*
// Use the C method to set the affinity of the thread to the CPU
#define _GNU_SOURCE
#include <sched.h>

void lock_thread(int cpuid) {
	cpu_set_t cpuset;

	CPU_ZERO(&cpuset);
	CPU_SET(cpuid, &cpuset);
	sched_setaffinity(0, sizeof(cpu_set_t), &cpuset);
}
*/
import "C"

// Default size of the array that will be populated by each thread.
const arrSize = (128 * 1024 * 1024)
const blockSize = 16

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func f(cpuID int, wg *sync.WaitGroup) {
	// Lock go thread with OS thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Signal workgroup when you are done!
	defer wg.Done()

	// Lock the OS thread with the CPU
	C.lock_thread(C.int(cpuID))

	// Stuff we need to stress the system
	outarr := make([]byte, 0, arrSize)
	var outfile *os.File
	var err error
	var i int = 0

	if outfile, err = ioutil.TempFile("", fmt.Sprintf("testfile-%s", strconv.Itoa(syscall.Gettid()))); err != nil {
		fmt.Sprintf("Unable to create tempfile :: %+v", err)
		os.Exit(1)
	}

	//Cleanup file on exit
	defer os.Remove(outfile.Name())

	// Do the work
	for i < arrSize {
		i = i + blockSize
		// If the below command is make([]byte, 0, 16), rand.Read never reads
		// anything into that array but creating it by nulling it from the
		// get-go, seems to be work.

		// Also, b has to be defined inside the loop. If defined
		// outside, rand.Read() returns empty again.
		b := make([]byte, 16)

		//Read some random data from the system's default random generator
		if _, err = rand.Read(b); err != nil {
			fmt.Println("Ouch!")
			return
		}

		// Fill the array and the tempfile with the data we just got.
		outarr = append(outarr, b...)
		if _, err = outfile.Write(b); err != nil {
			fmt.Sprintf("Unable to write to tempfile :: %+v", err)
			os.Exit(1)
		}
		b = nil
	}

	// Stats about the job
	fmt.Printf("PIDs : %+v : %+v: %+v\n", os.Getpid(), os.Getppid(), syscall.Gettid())
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
}

func main() {

	// Create workgroup to track all the workers.
	var wg sync.WaitGroup

	//Max number of CPUs.
	//TO-DO: Below line might not be needed as recent go release has this
	//set as default to be all CPUs.
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("CPUs: %+v\n", runtime.NumCPU())

	// Schedule job on all the CPUs
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go f(i, &wg)
	}

	wg.Wait()

	//Get Stats about the whole program
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}
