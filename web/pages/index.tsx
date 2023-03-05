import Link from 'next/link';

export default function Home() {
    return <>
        <header>
            <h1>Home</h1>
            <nav>
                <ul>
                    <li>
                        <Link href="/games">games</Link>
                    </li>
                </ul>
            </nav>
        </header>
    </>;
}
