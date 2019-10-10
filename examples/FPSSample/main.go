// Copyright 2017 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"agones.dev/agones/sdks/go"
)

type interceptor struct {
	forward   io.Writer
	intercept func(p []byte)
}

// Write will intercept the incoming stream, and forward
// the contents to its `forward` Writer.
func (i *interceptor) Write(p []byte) (n int, err error) {
	if i.intercept != nil {
		i.intercept(p)
	}

	return i.forward.Write(p)
}

// main intercepts the stdout of the game server and uses it
// to determine if the game server is ready or not.
func main() {
	input := flag.String("i", "", "path to server_linux.sh")
	readyText := flag.String("r", "", "server ready text")
	flag.Parse()

	fmt.Println(">>> Connecting to Agones with the SDK")
	s, err := sdk.NewSDK()
	if err != nil {
		log.Fatalf(">>> Could not connect to sdk: %v", err)
	}

	fmt.Println(">>> Starting health checking")
	go doHealth(s)

	fmt.Println(">>> Starting wrapper for game server!")
	fmt.Printf(">>> Path to game server script: %s \n", *input)

	// track references to listening count
	listeningCount := 0

	cmd := exec.Command(*input) // #nosec
	cmd.Stderr = &interceptor{forward: os.Stderr}
	cmd.Stdout = &interceptor{
		forward: os.Stdout,
		intercept: func(p []byte) {
			if listeningCount >= 4 {
				return
			}

			str := strings.TrimSpace(string(p))
			if count := strings.Count(str, *readyText); count > 0 {
				fmt.Printf(">>> Moving to READY: %s \n", str)
				err = s.Ready()
				if err != nil {
					log.Printf("Could not send ready message: %v", err)
				}
			}
		}}

	err = cmd.Start()
	if err != nil {
		log.Printf(">>> Error Starting Cmd %v", err)
	}
	err = cmd.Wait()
	log.Printf(">>> game server shutdown unexpectedly %v", err)
}

// doHealth sends the regular Health Pings
func doHealth(sdk *sdk.SDK) {
	tick := time.Tick(2 * time.Second)
	for {
		err := sdk.Health()
		if err != nil {
			log.Fatalf("[wrapper] Could not send health ping, %v", err)
		}
		<-tick
	}
}
