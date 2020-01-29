// Copyright 2019 HAProxy Technologies LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	c "github.com/haproxytech/kubernetes-ingress/controller"
	"github.com/jessevdk/go-flags"
)

func main() {

	c.FrontendHTTP = "http"
	c.FrontendHTTPS = "https"
	c.FrontendSSL = "ssl"
	c.HAProxyCFG = "/etc/haproxy/haproxy.cfg"
	c.HAProxyCertDir = "/etc/haproxy/certs/"
	c.HAProxyStateDir = "/var/state/haproxy/"

	var osArgs c.OSArgs
	var parser = flags.NewParser(&osArgs, flags.IgnoreUnknown)
	_, err := parser.Parse()
	exitCode := 0
	defer func() {
		os.Exit(exitCode)
	}()
	if err != nil {
		log.Println(err)
		exitCode = 1
		return
	}

	defaultBackendSvc := fmt.Sprintf("%s/%s", osArgs.DefaultBackendService.Namespace, osArgs.DefaultBackendService.Name)
	defaultCertificate := fmt.Sprintf("%s/%s", osArgs.DefaultBackendService.Namespace, osArgs.DefaultCertificate.Name)
	c.SetDefaultAnnotation("default-backend-service", defaultBackendSvc)
	c.SetDefaultAnnotation("ssl-certificate", defaultCertificate)

	if len(osArgs.Version) > 0 {
		fmt.Printf("HAProxy Ingress Controller %s %s%s\n\n", GitTag, GitCommit, GitDirty)
		fmt.Printf("Build from: %s\n", GitRepo)
		fmt.Printf("Build date: %s\n\n", BuildTime)
		if len(osArgs.Version) > 1 {
			fmt.Printf("ConfigMap: %s/%s\n", osArgs.ConfigMap.Namespace, osArgs.ConfigMap.Name)
			fmt.Printf("Ingress class: %s\n", osArgs.IngressClass)
		}
		return
	}

	if len(osArgs.Help) > 0 && osArgs.Help[0] {
		parser.WriteHelp(os.Stdout)
		return
	}

	log.Println(IngressControllerInfo)
	log.Printf("HAProxy Ingress Controller %s %s%s\n\n", GitTag, GitCommit, GitDirty)
	log.Printf("Build from: %s\n", GitRepo)
	log.Printf("Build date: %s\n\n", BuildTime)
	log.Printf("ConfigMap: %s/%s\n", osArgs.ConfigMap.Namespace, osArgs.ConfigMap.Name)
	log.Printf("Ingress class: %s\n", osArgs.IngressClass)
	if osArgs.ConfigMapTCPServices.Name != "" {
		log.Printf("TCP Services defined in %s/%s\n", osArgs.ConfigMapTCPServices.Namespace, osArgs.ConfigMapTCPServices.Name)
	}

	log.Printf("Default backend service: %s\n", defaultBackendSvc)
	log.Printf("Default ssl certificate: %s\n", defaultCertificate)

	ctx, cancel := context.WithCancel(context.Background())
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	go func() {
		<-channel
		cancel()
	}()

	if osArgs.Test {
		setupTestEnv()
	}

	hAProxyController := c.HAProxyController{}
	hAProxyController.Start(ctx, osArgs)
}
