def go_cmd = "/usr/local/go/bin/go"

node {
    checkout scm

    environment = "demo_env"

    stage('Run static analysis') {
        // TODO
    }

    stage('Run tests') {
        dir("test") {
            ansiColor('xterm') {
                sh "${go_cmd} test"
            }
        }
    }
}
