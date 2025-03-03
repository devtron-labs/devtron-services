/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("hello")
	app, err := InitializeApp()
	if err != nil {
		log.Panic(err)
	}
	go app.Start()
	//     gracefulStop start
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	sig := <-gracefulStop
	fmt.Printf("caught term sig: %+v", sig)
	app.Stop()
	//      gracefulStop end

	/*	cs := &klarService.KlarServiceImpl{}

		cs.Process(&common.ScanEvent{
			Image:        "686244538589.dkr.ecr.us-east-2.amazonaws.com/devtron:88775627-56-731",
			ImageDigest:  "sha256:49ad28c8b3b7b2485c4baf4d8538fd31cac4d9bd0aa25d120a8cede6d675ad7e",
			AppId:        0,
			EnvId:        0,
			PipelineId:   0,
			CiArtifactId: 0,
			UserId:       0,
		})*/

}
