import { useState, useCallback, useRef, useEffect } from "react";

export const useWebsocket = (url: string, callbackfn: (webSocket: WebSocket, parse: <T = any>(data: string) => T) => void, deps: any[]) => {
    const [connected, setConnected] = useState(false);
    const memofn = useCallback(callbackfn, deps);
    const webSocket = useRef<WebSocket>();

    useEffect(() => {
        const current = new WebSocket(url);

        callbackfn(current, function parse<T = any>(data: string) { return JSON.parse(data) as T; });

        webSocket.current = current;

        return () => {
            current.close();
        };
    }, [url, memofn]);


    function send(data: Record<string, any>) { webSocket.current?.send(JSON.stringify(data)); }

    return [send, connected] as [send: typeof send, connected: boolean];
};