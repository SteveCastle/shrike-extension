import browser from "webextension-polyfill";

type CommandData = { Command: string; Args: string[] };

const runCommand = (data: CommandData) => {
    async function callCommandServer() {
        const response = await fetch("http://localhost:8090", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(data),
        });
        const text = await response.text();
        console.log(text);
    }
    callCommandServer();
};

browser.runtime.onMessage.addListener(
    (request: { message: string; data: CommandData }) => {
        console.log("running Listener");
        console.log(request.data);
        switch (request.message) {
            case "runCommand": {
                runCommand(request.data);
                break;
            }
            default:
        }
    },
);
