
export function createCurlCommand(subCommand, data) {
    return `curl --request PUT \\\n  --url ${window.location.hostname}/api${subCommand} \\\n  ` +
        `--header 'content-type: application/json' \\\n  --data '${JSON.stringify(data)}'`
}
