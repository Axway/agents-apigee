#!/bin/bash

sonar-scanner -X \
    -Dsonar.host.url=${SONAR_HOST} \
    -Dsonar.language=go \
    -Dsonar.projectName=Apigee_Agent_Client \
    -Dsonar.projectVersion=1.0 \
    -Dsonar.projectKey=Apigee_Agent_Client \
    -Dsonar.sourceEncoding=UTF-8 \
    -Dsonar.projectBaseDir=${WORKSPACE} \
    -Dsonar.sources=./client/pkg/apigee/*,./client/pkg/apigee/models/*,./client/pkg/apigee/cmd,./discovery/** \
    -Dsonar.tests=./client/pkg/**,./discovery/**,./traceability/** \
	-Dsonar.exclusions=**/*.json \
    -Dsonar.test.inclusions=**/*test*.go \
    -Dsonar.go.tests.reportPaths=goreport.json \
    -Dsonar.go.coverage.reportPaths=gocoverage.out \
    -Dsonar.issuesReport.console.enable=true \
    -Dsonar.report.export.path=sonar-report.json \
    -Dsonar.verbos=true
