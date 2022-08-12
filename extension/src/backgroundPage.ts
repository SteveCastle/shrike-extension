import browser from "webextension-polyfill";

// Listen for messages sent from other parts of the extension
browser.runtime.onMessage.addListener((request: { popupMounted: boolean }) => {
    // Log statement if request.popupMounted is true
    // NOTE: this request is sent in `popup/component.tsx`
    if (request.popupMounted) {
        console.log("backgroundPage notified that Popup.tsx has mounted.");
    }
});

const runCommand = (data: string) => {
    async function callCommandServer() {
        const results = await fetch("http://localhost:8090", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: data,
        });
        const json = await results.json();
        console.log(json);
    }
    callCommandServer();
};

browser.runtime.onMessage.addListener(
    (request: { message: string; data: string }) => {
        console.log(request);
        switch (request.message) {
            case "save": {
                runCommand(request.data);
                break;
            }
            default:
        }
    },
);
