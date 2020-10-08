async function json(method, url, data, json, headers) {
    const options = {
        method: method,
        body: json != null ? JSON.stringify(json) : data,
        mode: "cors",
        cache: "no-cache",
        credentials: "same-origin",
        headers: headers || {
            "Content-Type": "application/json"
        },
        redirect: "follow",
        referrerPolicy: "no-referrer",
    };
    const response = await fetch(url, options);
    const result = {
        statusCode: response.status,
        status: `${response.status} ${response.statusText}`,
        json: await response.json()
    };

    if (response.status >= 200 && response.status < 300) {
        return result;
    } else {
        return Promise.reject(result);
    }
}

async function download(url, headers) {

    const options = {
        mode: "cors",
        cache: "no-cache",
        credentials: "same-origin",
        headers: headers || {},
        redirect: "follow",
        referrerPolicy: "no-referrer",
    };
    const response = await fetch(url, options);
    const result = {
        statusCode: response.status,
        status: `${response.status} ${response.statusText}`,
        blob: await response.blob()
    };

    if (response.status >= 200 && response.status < 300) {
        return result;
    } else {
        return Promise.reject(result);
    }
}

const backend = {
    get: (url = "", headers = null) =>
        json("GET", url, null, null, headers),
    post: (url = "", data = null, headers = null) =>
        json("POST", url, null, data, headers),
    put: (url = "", data = null, headers = null) =>
        json("PUT", url, null, data, headers),
    delete: (url = "", data = null, headers = null) =>
        json("DELETE", url, null, data, headers),
    postFile: (url = "", data = null, headers = {}) =>
        json("POST", url, data, null, headers),
    download: (url = "", headers = null) =>
        download(url, headers)
};

export default backend;
