
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
    const result = await fetch(url, options);
    return result.json();
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
    postFile: (url = "", data = null, headers = null) => {
        return file(url, data, headers);
    },
};

export default backend;
