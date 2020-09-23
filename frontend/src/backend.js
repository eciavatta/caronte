
async function request(method, url, data) {
    const options = {
        method: method,
        mode: "cors",
        cache: "no-cache",
        credentials: "same-origin",
        headers: {
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

const backend = {
    get: (url = "") => {
        return request("GET", url, null);
    },
    post: (url = "", data = null) => {
        return request("POST", url, data);
    },
    put: (url = "", data = null) => {
        return request("PUT", url, data);
    },
    delete: (url = "", data = null) => {
        return request("DELETE", url, data);
    }
};

export default backend;
