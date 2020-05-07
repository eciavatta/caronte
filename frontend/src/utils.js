export function createCurlCommand(subCommand, data) {
    let full = window.location.protocol + '//' + window.location.hostname + (window.location.port ? ':' + window.location.port : '');
    return `curl --request PUT \\\n  --url ${full}/api${subCommand} \\\n  ` +
        `--header 'content-type: application/json' \\\n  --data '${JSON.stringify(data)}'`;
}
