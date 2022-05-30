#!/usr/bin/env groovy
@Library('apm@current') _
pipeline {
  agent { label 'linux && immutable && debian-11' }
  environment {
    VAULT_ADDR=credentials('vault-addr')
    VAULT_ROLE_ID=credentials('apm-vault-role-id')
    VAULT_SECRET_ID=credentials('apm-vault-secret-id')
    REPO = 'cloudbeat'
    BASE_DIR = "src/github.com/elastic/${env.REPO}"
    JOB_GCS_BUCKET = 'internal-ci-artifacts'
    JOB_GCS_CREDENTIALS = 'internal-ci-gcs-plugin'
    DIAGNOSTIC_INTERVAL = "${params.DIAGNOSTIC_INTERVAL}"
    ES_LOG_LEVEL = "${params.ES_LOG_LEVEL}"
    DOCKER_SECRET = 'secret/apm-team/ci/docker-registry/prod'
    DOCKER_REGISTRY = 'docker.elastic.co'
    WORKFLOW = "${params.Build_type}"
  }
  options {
    timeout(time: 2, unit: 'HOURS')
    buildDiscarder(logRotator(numToKeepStr: '100', artifactNumToKeepStr: '30', daysToKeepStr: '30'))
    timestamps()
    ansiColor('xterm')
    disableResume()
    durabilityHint('PERFORMANCE_OPTIMIZED')
    rateLimitBuilds(throttle: [count: 60, durationName: 'hour', userBoost: true])
    quietPeriod(10)
  }
  parameters {
    booleanParam(name: 'Run_As_Main_Branch', defaultValue: false, description: 'Allow to run any steps on a PR, some steps normally only run on main branch.')
    booleanParam(name: 'linux_ci', defaultValue: true, description: 'Enable Linux build')
    booleanParam(name: 'intake_ci', defaultValue: true, description: 'Enable test')
    booleanParam(name: 'test_ci', defaultValue: false, description: 'Enable test')
    booleanParam(name: 'release_ci', defaultValue: true, description: 'Enable build the release packages')
    booleanParam(name: 'its_ci', defaultValue: true, description: 'Enable async ITs')
    choice(name: 'Build_type', choices: ['snapshot', 'staging'], description: 'Choose Snapshot or Staging Build type(Default: Snapshot)')
    string(name: 'DIAGNOSTIC_INTERVAL', defaultValue: "0", description: 'Elasticsearch detailed logging every X seconds')
    string(name: 'ES_LOG_LEVEL', defaultValue: "error", description: 'Elasticsearch error level')
  }
  stages {
    /**
    Checkout the code and stash it, to use it on other stages.
    */
    stage('Checkout') {
      environment {
        PATH = "${env.PATH}:${env.WORKSPACE}/bin"
        HOME = "${env.WORKSPACE}"
      }
      options { skipDefaultCheckout() }
      steps {
        // pipelineManager([ cancelPreviousRunningBuilds: [ when: 'PR' ] ])
        deleteDir()
        gitCheckout(basedir: "${BASE_DIR}", githubNotifyFirstTimeContributor: true,
                    shallow: false, reference: "/var/lib/jenkins/.git-references/${REPO}.git")
        stash allowEmpty: true, name: 'source', useDefaultExcludes: false
      }
    }
    /**
    Updating generated files for Beat.
    Checks the GO environment.
    Checks the Python environment.
    Checks YAML files are generated.
    Validate that all updates were committed.
    */
    stage('Intake') {
      options { skipDefaultCheckout() }
      environment {
        PATH = "${env.PATH}:${env.WORKSPACE}/bin"
        HOME = "${env.WORKSPACE}"
      }
      when {
        beforeAgent true
        allOf {
          expression { return params.intake_ci }
          // expression { return env.ONLY_DOCS == "false" }
        }
      }
      steps {
        // withGithubNotify(context: 'Intake') {
          deleteDir()
          unstash 'source'
          dir("${BASE_DIR}"){
            withGoEnv(){
              sh(label: 'Run intake', script: './.ci/scripts/intake.sh')
            }
          }
        }
      // }
    }
    stage('Build and Test'){
      failFast false
      parallel {
        /**
        Build on a linux environment.
        */
        stage('linux build') {
          options { skipDefaultCheckout() }
          when {
            beforeAgent true
            allOf {
              expression { return params.linux_ci }
            }
          }
          steps {
            deleteDir()
            unstash 'source'
            dir(BASE_DIR){
              withMageEnv(){
                retry(2) { // Retry in case there are any errors to avoid temporary glitches
                  sleep randomNumber(min: 5, max: 10)
                  sh(label: 'Linux build', script: './.ci/scripts/build.sh')
                }
              }
            }
          }
        }
        /**
          Run unit tests and report junit results.
        */
        stage('Unit Test') {
          agent { label 'linux && immutable && debian-11' }
          options { skipDefaultCheckout() }
          environment {
            PATH = "${env.PATH}:${env.WORKSPACE}/bin"
            HOME = "${env.WORKSPACE}"
            TEST_COVERAGE = "true"
          }
          when {
            beforeAgent true
            allOf {
              expression { return params.test_ci }
            }
          }
          steps {
              deleteDir()
              unstash 'source'
              dir("${BASE_DIR}"){
                withMageEnv(){
                  sh(label: 'Run Unit tests', script: './.ci/scripts/unit-test.sh')
                }
              }
          }
        }
        /**
        Packages Artifacts & Publishes release
        */
        stage('Package&Publish') {
          agent { label 'linux && immutable && debian-11' }
          options { skipDefaultCheckout() }
          environment {
            PATH = "${env.PATH}:${env.WORKSPACE}/bin"
            HOME = "${env.WORKSPACE}"
          }
          when {
            beforeAgent true
            allOf {
              expression { return params.release_ci }
              anyOf {
                branch 'main'
                branch pattern: '\\d+\\.\\d+', comparator: 'REGEXP'
                tag pattern: 'v\\d+\\.\\d+\\.\\d+.*', comparator: 'REGEXP'
                expression { return params.Run_As_Main_Branch }
              }
            }
          }
          stages {
            stage('Package') {
              steps {
                withGithubNotify(context: 'Package') {
                  deleteDir()
                  unstash 'source'
                  
                  dir("${BASE_DIR}"){
                    withMageEnv(){
                      sh(label: 'Build packages', script: './.ci/scripts/package.sh')
                    }
                  }
                }
              }
            }
            stage('Publish') {
              environment {
                BUCKET_URI = """${isPR() ? "gs://${JOB_GCS_BUCKET}/cloudbeat/pull-requests/pr-${env.CHANGE_ID}" : "gs://${JOB_GCS_BUCKET}/cloudbeat/snapshots"}"""
              }
              steps {
                // Login to Docker Registery
                dockerLogin(secret: env.DOCKER_SECRET, registry: env.DOCKER_REGISTRY)
                
                // Upload files to the default location
                googleStorageUpload(bucket: "${BUCKET_URI}",
                  credentialsId: "${JOB_GCS_CREDENTIALS}",
                  pathPrefix: "${BASE_DIR}/build/distributions/",
                  pattern: "${BASE_DIR}/build/distributions/**/*",
                  sharedPublicly: true,
                  showInline: true)
                  
                  // Call rm-docker command
                  dir("${BASE_DIR}"){
                  sh(label: 'Release-manager-docker', script: './.ci/scripts/rm-docker.sh')
                  }
              }
            }
          } // Package&Publish stages
        } // Package&Publish
      } // build&test stages
    } // build&test
  } // stages
}