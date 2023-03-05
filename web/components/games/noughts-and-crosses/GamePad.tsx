import { ButtonHTMLAttributes } from "react";
import { CSS } from "./styles";

type Value = 0 | 1 | 2;

type GamePadProps = {
} & ButtonHTMLAttributes<HTMLButtonElement>;

export function GamePad({ disabled, ...props }: GamePadProps) {
    // if a client (1|2) has made a move, disable the button
    return <button {...props} disabled={disabled || props.value !== 0} style={CSS.pad}>{convert(props.value as Value)}</button>;
}

const convert = (value: Value) => {
    switch (value) {
        case 0:
            return "-";
        case 1:
            return "X";
        case 2:
            return "O";
        default:
            return " ";
    }
};