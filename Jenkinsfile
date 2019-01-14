notifyBuild('STARTED')

node {
  try {
    githubNotify(status: 'PENDING')

    docker.image('behrsin/go-node').inside('-u root') {
      stage('checkout') {
        dir('iot-backend') {
          checkout scm
        }

        dir('go-v8') {
          git branch: 'master', url: 'https://github.com/behrsin/go-v8.git'
        }
      }

      dir('iot-backend') {
        stage('deps') {
          sh 'apt-get update'
          sh 'apt-get -y install protobuf-compiler'
          sh 'go version'
          sh 'make deps'
        }

        stage('build') {
          sh 'make'
        }

        stage('test') {
          sh 'make test'
        }
      }
    }
  } catch (e) {
    currentBuild.result = "FAILED"
    githubNotify(status: 'ERROR')
  } finally {
    notifyBuild(currentBuild.result)

    def status = currentBuild.result ?: 'SUCCESSFUL'
    if(status == 'SUCCESSFUL'){
      githubNotify(status: 'SUCCESS')
    }
  }
}

def notifyBuild(String buildStatus = 'STARTED') {
  buildStatus =  buildStatus ?: 'SUCCESSFUL'

  def colorCode = '#cc2626'
  def subject = "${buildStatus}: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'"
  def summary = "${subject} <${env.BUILD_URL}|Job URL> - <${env.BUILD_URL}/console|Console Output>"

  if (buildStatus == 'STARTED') {
    colorCode = '#ff9900'
  } else if (buildStatus == 'SUCCESSFUL') {
    colorCode = '#1ec41e'
  } else {
    colorCode = '#cc2626'
  }

  slackSend(color: colorCode, message: summary)
}
