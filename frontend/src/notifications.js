import log from "./log";
import dispatcher from "./dispatcher";

class Notifications {

    constructor() {
        const location = document.location;
        this.wsUrl = `ws://${location.hostname}${location.port ? ":" + location.port : ""}/ws`;
    }

    createWebsocket = () => {
        this.ws = new WebSocket(this.wsUrl);
        this.ws.onopen = this.onWebsocketOpen;
        this.ws.onerror = this.onWebsocketError;
        this.ws.onclose = this.onWebsocketClose;
        this.ws.onmessage = this.onWebsocketMessage;
    };

    onWebsocketOpen = () => {
        log.debug("Connected to backend with websocket");
    };

    onWebsocketError = (err) => {
        this.ws.close();
        log.error("Websocket error", err);
        setTimeout(() => this.createWebsocket(), 3000);
    };

    onWebsocketClose = () => {
        log.debug("Closed websocket connection with backend");
    };

    onWebsocketMessage = (message) => {
        dispatcher.dispatch("notifications", JSON.parse(message.data));
    };
}

const notifications = new Notifications();

export default notifications;
