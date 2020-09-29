
async function json(method, url, data, headers) {
    const options = {
        method: method,
        mode: "cors",
        cache: "no-cache",
        credentials: "same-origin",
        headers: headers || {
            "Content-Type": "application/json"
        },
        redirect: "follow",
        referrerPolicy: "no-referrer",
    };
    if (data != null) {
        options.body = JSON.stringify(data);
    }
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

async function file(url, data, headers) {
    const options = {
        method: "POST",
        mode: "cors",
        cache: "no-cache",
        credentials: "same-origin",
        body: data,
        redirect: "follow",
        referrerPolicy: "no-referrer",
    };
    return await fetch(url, options);
}

const backend = {
    get: (url = "", headers = null) => {
        return json("GET", url, null, headers);
    },
    post: (url = "", data = null, headers = null) => {
        return json("POST", url, data, headers);
    },
    put: (url = "", data = null, headers = null) => {
        return json("PUT", url, data, headers);
    },
    delete: (url = "", data = null, headers = null) => {
        return json("DELETE", url, data, headers);
    },
    getJson: (url = "", headers = null) => {
        return json("GET", url, null, headers);
    },
    postJson: (url = "", data = null, headers = null) => {
        return json("POST", url, data, headers);
    },
    putJson: (url = "", data = null, headers = null) => {
        return json("PUT", url, data, headers);
    },
    deleteJson: (url = "", data = null, headers = null) => {
        return json("DELETE", url, data, headers);
    },
    postFile: (url = "", data = null, headers = null) => {
        return file(url, data, headers);
    },
};

export default backend;
