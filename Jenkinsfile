try {
  notifyBuild('STARTED')
  githubNotify(status: 'PENDING')

  pipeline {
    agent {
      docker 'golang:1.11-stretch'
    }

    stages {
      ws("${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}/") {
        withEnv(["GOPATH=${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}"]) {
          env.PATH="${GOPATH}/bin:$PATH"

          stage('Checkout') {
            echo 'Checking out SCM'
            checkout scm
          }

          stage('Pre Test') {
            echo 'Pulling Dependencies'

            sh 'go version'
            sh 'make deps'
          }

          stage('Test') {
            sh 'make test'
          }

          stage('Build') {
            sh 'make'
          }
        }
      }
    }
  }
} catch (e) {
  // If there was an exception thrown, the build failed
  currentBuild.result = "FAILED"

  githubNotify(status: 'ERROR')

} finally {
  // Success or failure, always send notifications
  notifyBuild(currentBuild.result)

  def bs = currentBuild.result ?: 'SUCCESSFUL'
  if(bs == 'SUCCESSFUL'){
    githubNotify(status: 'SUCCESS')
  }
}

def notifyBuild(String buildStatus = 'STARTED') {
  // build status of null means successful
  buildStatus =  buildStatus ?: 'SUCCESSFUL'

  // Default values
  def colorName = 'RED'
  def colorCode = '#FF0000'
  def subject = "${buildStatus}: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'"
  def summary = "${subject} <${env.BUILD_URL}|Job URL> - <${env.BUILD_URL}/console|Console Output>"

  // Override default values based on build status
  if (buildStatus == 'STARTED') {
    color = 'YELLOW'
    colorCode = '#FFFF00'
  } else if (buildStatus == 'SUCCESSFUL') {
    color = 'GREEN'
    colorCode = '#00FF00'
  } else {
    color = 'RED'
    colorCode = '#FF0000'
  }

  // Send notifications
  slackSend(color: colorCode, message: summary)
}
