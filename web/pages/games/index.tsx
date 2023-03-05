import Link from "next/link";

export default function Page() {
    return <>
        <header>
            <h1>Games</h1>
            <nav>
                <ul>
                    <li>
                        <Link href="/">home</Link>
                    </li>
                    <li>
                        <Link href="/games/noughts-and-crosses">noughts and crosses</Link>
                    </li>
                </ul>
            </nav>
        </header>
    </>;
}