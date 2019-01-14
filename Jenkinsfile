notifyBuild('STARTED')

node {
  try {
    githubNotify(status: 'PENDING')

    docker.image('golang:1.11-stretch').inside() {
      stage('Checkout') {
        echo 'Checking out SCM'

        dir('iot-backend') {
          checkout scm
        }

        dir('go-v8') {
          git branch: 'master', url: 'https://github.com/behrsin/go-v8.git'
        }
      }

      dir('iot-backend') {
        stage('Pre Test') {
          echo 'Pulling Dependencies'

          sh 'go version'
          sh 'make deps'
        }

        stage('Build') {
          sh 'make'
        }

        stage('Test') {
          sh 'make test'
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
    colorCode = '#ff9900'
  } else if (buildStatus == 'SUCCESSFUL') {
    colorCode = '#1ec41e'
  } else {
    colorCode = '#cc2626'
  }

  // Send notifications
  slackSend(color: colorCode, message: summary)
}
