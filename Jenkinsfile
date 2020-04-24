//
// Copyright (c) 2020 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

pipeline {
    agent { label 'centos7-docker-4c-2g' }
    options {
        timestamps()
    }
    stages {
        stage('Prep Base Build Image') {
            steps {
                script { docker.build('edgex-go-ci-base', '-f Dockerfile.build .') }
                sh 'docker save -o base.tar edgex-go-ci-base'
                stash name: 'ci-base', includes: '**/base.tar'
            }
        }

        stage('Parallel Docker') {
            agent { label 'centos7-docker-4c-2g' }
            environment {
                BUILDER_BASE = 'edgex-go-ci-base'
            }
            steps {
                unstash 'ci-base'

                sh 'docker import base.tar $BUILDER_BASE'
                sh 'rm -rf base.tar'

                script {
                    sh 'sudo curl -L "https://github.com/docker/compose/releases/download/1.25.5/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose'
                    sh 'sudo chmod +x /usr/local/bin/docker-compose'

                    def dockers = [
                        [image: 'core-metadata-go', dockerfile: 'cmd/core-metadata/Dockerfile'],
                        [image: 'core-data-go', dockerfile: 'cmd/core-data/Dockerfile'],
                        [image: 'core-command-go', dockerfile: 'cmd/core-command/Dockerfile'],
                        [image: 'support-logging-go', dockerfile: 'cmd/support-logging/Dockerfile'],
                        [image: 'support-notifications-go', dockerfile: 'cmd/support-notifications/Dockerfile'],
                        [image: 'support-scheduler-go', dockerfile: 'cmd/support-scheduler/Dockerfile'],
                        [image: 'sys-mgmt-agent-go', dockerfile: 'cmd/sys-mgmt-agent/Dockerfile'],
                        [image: 'edgex-secrets-setup-go', dockerfile: 'cmd/security-secrets-setup/Dockerfile'],
                        [image: 'edgex-security-proxy-setup-go', dockerfile: 'cmd/security-proxy-setup/Dockerfile'],
                        [image: 'edgex-security-secretstore-setup-go', dockerfile: 'cmd/security-secretstore-setup/Dockerfile'
                    ]

                    def dockerCompose = generateDockerComposeForBuild(dockers)
                    writeFile file: 'docker-compose.yml', text: dockerCompose
                    sh 'cat docker-compose.yml'
                    sh 'docker-compose build --parallel'
                }
            }
        }

        // stage('make docker with cache') {
        //     agent { label 'centos7-docker-4c-2g' }
        //     environment {
        //         BUILDER_BASE = 'edgex-go-ci-base'
        //     }
        //     steps {
        //         unstash 'ci-base'
        //         sh 'docker import base.tar $BUILDER_BASE'
        //         sh 'rm -rf base.tar'
        //         sh 'make docker'
        //     }
        // }

        // stage('Current make docker') {
        //     agent { label 'centos7-docker-4c-2g' }
        //     steps {
        //          sh 'make docker'
        //     }
        // }
    }
}

def generateDockerComposeForBuild(services) {
"""
version: '3'

  services:
    ${services.collect { generateServiceYaml(it.image, it.dockerfile, env.GIT_COMMIT) }.join('\n') }
"""
}

def generateServiceYaml(serviceName, dockerFile, gitCommit) {
"""
  ${serviceName}:
    build:
      context: .
      dockerfile: ${dockerFile}
      labels:
        - git_sha=${gitCommit}
      args:
        - BUILDER_BASE
    image: edgexfoundry/docker-${serviceName}"""
}
