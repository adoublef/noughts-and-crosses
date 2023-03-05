import { CSSProperties } from "react";

export const CSS: Record<"board" | "pad", CSSProperties> = {
    board: {
        display: "grid",
        gap: "1px",
        gridTemplateColumns: "repeat(3, 1fr)",
        width: "max-content"
    },
    pad: {
        height: "100px",
        width: "100px",
    }
};