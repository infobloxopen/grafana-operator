
@Library('jenkins.shared.library') _ 

pipeline {
  agent {
    label 'ubuntu_docker_label'
  }
  tools {
    go "Go 1.16.3"
  }
  options {
    checkoutToSubdirectory('grafana-operator')
  }
  stages {
    stage("Setup") {
      steps {
        prepareBuild()
      }
    }
    stage("Build Image") {
      steps {
        dir("grafana-operator") {
          sh "make setup/travis image/build"
        }
      }
    }
  }
  post {
    success {
      finalizeBuild(
        sh(
          script: 'make image/show',
          returnStdout: true
        )
      )
    }
  }
}
