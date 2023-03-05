import { useWebsocket } from "@/lib/websockets";
import { ButtonHTMLAttributes, HTMLAttributes, MouseEventHandler, useCallback, useEffect, useRef, useState } from "react";
import { GamePad } from "./GamePad";
import { CSS } from "./styles";

type Value = 0 | 1 | 2;

type Board = [
    Value, Value, Value,
    Value, Value, Value,
    Value, Value, Value
];

type Client = {
    id: string;
    value: Exclude<Value, 0>;
};

export type GameBoardProps = {
    connected: boolean;
    board: Board;
    current: Client;
    clients: [Client] | [Client, Client];
    onMove: (value: Value, index: number) => MouseEventHandler<HTMLButtonElement>;
    onReset: () => void;
} & ButtonHTMLAttributes<HTMLButtonElement>;

export function GameBoard(props: GameBoardProps) {
    const winner = combinations.reduce<0 | 1 | 2>(evaluate(props.board), 0);

    // disable if there are not 2 clients or if the board is full or if the game is over
    const disabled = props.clients.length !== 2 || props.board.every(value => value !== 0) || !!winner;

    return (<>
        {props.clients.length !== 2 && <p>waiting for another player</p>}
        {!!winner && <p>Congratulations: {props.clients[Number(winner !== props.clients[0].value)].id}</p>}
        <div className="board" style={CSS.board}>
            {props.board.map((value, index) => (
                <GamePad key={index} onMouseDown={props.onMove(value, index)} disabled={disabled} value={value} />
            ))}
        </div>
        <div className="status">
            <p>current: {props.current.id}</p>
        </div>
        <button className="reset" onClick={props.onReset}>reset</button>
    </>);
}

/** this will be generated during onconnect */
const initState = {
    board: [
        0, 0, 0,
        0, 0, 0,
        0, 0, 0,
    ],
    current: { id: "foo", value: 1 },
    clients: [
        { id: "foo", value: 1 },
        { id: "bar", value: 2 },
    ],
} satisfies GameBoardState;

type GameBoardState = {
    board: Board;
    current: Client;
    clients: [Client] | [Client, Client];
};

export const useGameBoard = (url: string) => {
    const [current, setCurrent] = useState(initState.clients[0] as Client);
    const [board, setBoard] = useState(initState.board as Board);
    const [clients, _setClients] = useState(initState.clients as [Client] | [Client, Client]);

    const [send, connected] = useWebsocket(url, (ws, parse) => {
        ws.onopen = () => {
            send({ type: "connect" });
        };

        ws.onmessage = event => {
            const message = parse<{ type: string, payload: any; }>(event.data);

            switch (message.type) {
                case "connect":
                    console.log("connected");
                    break;
                case "move":
                    console.log("move", message.payload);
                    break;
                case "reset":
                    setBoard(message.payload.board);
                    break;
                default:
                    console.log("unknown message", message);
            }
        };

        ws.onclose = () => console.log("disconnected");
    }, []);


    function onMove(_value: number, index: number): MouseEventHandler<HTMLButtonElement> {
        return () => {
            const move = {
                type: "move",
                payload: {
                    index,
                    value: current.value,
                },
            };

            send(move);

            setBoard(prev => prev.map((pValue, pIndex) => {
                if (pIndex !== index) return pValue;
                // NOTE -- this is the same as using clients.find but without returning undefined
                setCurrent(prev => clients[Number(prev.id === initState.clients[0].id)]);
                return current.value;
            }) as Board);
        };// NOTE -- a map is 1:1 so this is safe
    };

    function onReset() {
        send({ type: "reset", payload: { board: [0, 0, 0, 0, 0, 0, 0, 0, 0] } });
    }

    // hoisting for webSocket.onmessage, webSocket.onopen, webSocket.onclose to use state
    function onOpen() { console.log("connected"); }

    function onMessage(event: MessageEvent) { console.log("message", event.data); }

    function onClose() { console.log("disconnected"); }

    return { connected, current, clients, board, onMove, onReset } satisfies GameBoardProps;
};

const combinations = [
    [0, 1, 2],
    [3, 4, 5],
    [6, 7, 8],
    [0, 3, 6],
    [1, 4, 7],
    [2, 5, 8],
    [0, 4, 8],
    [2, 4, 6],
] as const;

type Combination = typeof combinations[number];

const evaluate = (board: Board) => (acc: Value, [a, b, c]: Combination) => {
    if (acc !== 0) return acc;
    return (board[a] === board[b] && board[a] === board[c]) ? board[a] : acc;
};

