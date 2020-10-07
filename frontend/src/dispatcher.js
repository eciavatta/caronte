
class Dispatcher {

    constructor() {
        this.listeners = [];
    }

    dispatch = (topic, payload) => {
        this.listeners.filter(l => l.topic === topic).forEach(l => l.callback(payload));
    };

    register = (topic, callback) => {
        if (typeof callback !== "function") {
            throw new Error("dispatcher callback must be a function");
        }
        if (typeof topic === "string") {
            this.listeners.push({topic, callback});
        } else if (typeof topic === "object" && Array.isArray(topic)) {
            topic.forEach(e => {
                if (typeof e !== "string") {
                    throw new Error("all topics must be strings");
                }
            });

            topic.forEach(e => this.listeners.push({e, callback}));
        } else {
            throw new Error("topic must be a string or an array of strings");
        }
    };

}

const dispatcher = new Dispatcher();

export default dispatcher;
