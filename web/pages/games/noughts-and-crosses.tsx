import { GameBoard, GameBoardProps, useGameBoard } from "@/components/games/noughts-and-crosses/GameBoard";
import { GetServerSideProps, InferGetServerSidePropsType } from "next";
import { useEffect, useState } from "react";

export const getServerSideProps: GetServerSideProps = async (context) => {
    const wsUrl = `ws://localhost:8080/games/play`;

    return {
        props: {
            wsUrl,
        },
    };
};


export default function Page(props: InferGetServerSidePropsType<typeof getServerSideProps>) {
    const state = useGameBoard(props.wsUrl as string);

    return <>
        <header>
            <h1>noughts and crosses</h1>
        </header>
        <main>
            <h1>Let's play</h1>
            <GameBoard {...state} />
        </main>
    </>;
}