import React, { useEffect } from "react";
import css from "./styles.module.css";
import { browser, Tabs } from "webextension-polyfill-ts";
import { useLocalStorage } from "../../useLocalStorage";

function download(command: string, args: string[]): void {
    // Query for the active tab in the current window
    browser.tabs
        .query({ active: true, currentWindow: true })
        .then((tabs: Tabs.Tab[]) => {
            // Pulls current tab from browser.tabs.query response
            const currentTab: Tabs.Tab | undefined = tabs[0];
            currentTab.url;
            // Short circuits function execution is current tab isn't found
            if (!currentTab) {
                return;
            }
            chrome.runtime.sendMessage({
                message: "runCommand",
                data: {
                    Command: command,
                    Arguments: [...args, currentTab.url],
                },
            });
        });
}

type EditorTypes =
    | { key: "COMMAND" }
    | { key: "ARGUMENT"; index: number }
    | { key: "CREATE_ARGUMENT" }
    | false;

export function Command() {
    const [showEditor, setShowEditor] = React.useState<EditorTypes>(false);
    const [editorValue, setEditorValue] = React.useState<string>("");
    const [command, setCommand] = useLocalStorage<string>("command", "curl");
    const [args, setArgs] = useLocalStorage<string[]>("args", ["-X", "GET"]);

    const [tab, setTab] = React.useState<Tabs.Tab | undefined>(undefined);
    useEffect(() => {
        async function getTab() {
            const tabs = await browser.tabs.query({
                active: true,
                currentWindow: true,
            });
            const currentTab = tabs[0];
            if (!currentTab) {
                return;
            }
            setTab(currentTab);
        }
        getTab();
    }, []);
    return !showEditor ? (
        <div className={css.container}>
            <div>
                <div
                    className={css.command}
                    onClick={() => {
                        setEditorValue(command);
                        setShowEditor({ key: "COMMAND" });
                    }}
                >
                    {command}
                </div>
                {args.map((arg, index) => (
                    <div
                        key={index}
                        className={css.arg}
                        onClick={() => {
                            setEditorValue(arg);
                            setShowEditor({ key: "ARGUMENT", index });
                        }}
                    >
                        {arg}
                    </div>
                ))}
                <button
                    className={css.addArg}
                    onClick={() => {
                        setEditorValue("");
                        setShowEditor({ key: "CREATE_ARGUMENT" });
                    }}
                >
                    +
                </button>
                <div className={css.url}>URL</div>
            </div>

            <div className={css.btnContainer}>
                <button
                    className={css.btn}
                    onClick={() => {
                        download(command, args);
                    }}
                >
                    RUN
                </button>
            </div>
        </div>
    ) : (
        <div className={css.editorContainer}>
            <div className={css.editor}>
                <input
                    type="text"
                    value={editorValue}
                    onChange={(e) => setEditorValue(e.target.value)}
                />
                {showEditor.key === "ARGUMENT" && (
                    <button
                        className={css.removeButton}
                        onClick={() => {
                            setArgs((args) => {
                                args.splice(showEditor.index, 1);
                                return args;
                            });
                            setShowEditor(false);
                        }}
                    >
                        Remove
                    </button>
                )}
                <button
                    className={css.saveButton}
                    onClick={() => {
                        if (showEditor.key === "CREATE_ARGUMENT") {
                            if (editorValue.length > 0) {
                                setArgs((args) => [...args, editorValue]);
                            }
                        } else if (showEditor.key === "ARGUMENT") {
                            setArgs((args) => {
                                args[showEditor.index] = editorValue;
                                return args;
                            });
                        } else if (showEditor.key === "COMMAND") {
                            setCommand(editorValue);
                        }
                        setShowEditor(false);
                    }}
                >
                    Submit
                </button>
            </div>
        </div>
    );
}
