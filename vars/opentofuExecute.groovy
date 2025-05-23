import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/opentofuExecute.yaml'

void call(Map parameters = [:]) {
    List credentials = [[type: 'file', id: 'opentofuSecrets', env: ['PIPER_opentofuSecrets']]]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}