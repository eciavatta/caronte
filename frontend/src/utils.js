export function createCurlCommand(subCommand, data) {
    let full = window.location.protocol + '//' + window.location.hostname + (window.location.port ? ':' + window.location.port : '');
    return `curl --request PUT \\\n  --url ${full}/api${subCommand} \\\n  ` +
        `--header 'content-type: application/json' \\\n  --data '${JSON.stringify(data)}'`;
}

export function objectToQueryString(obj) {
    let str = [];
    for (let p in obj)
        if (obj.hasOwnProperty(p)) {
            str.push(encodeURIComponent(p) + "=" + encodeURIComponent(obj[p]));
        }
    return str.join("&");
}
