import React from "react";
import { Command } from "@src/components/command";
import css from "./styles.module.css";

export function Popup() {
    return (
        <div className={css.popupContainer}>
            <div className="mx-4 my-4">
                <Command />
            </div>
        </div>
    );
}
