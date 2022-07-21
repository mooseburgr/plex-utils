import { Html, Head, Main, NextScript } from 'next/document'

export default function Document() {
    return (
        <Html>
            <Head>
                <meta name="description" content="invite yourself to my plex server" />
                <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no"></meta>
                <link rel="icon" href="https://avatars.slack-edge.com/2022-01-11/2950060844657_4cae9e95e482718f4ef6_88.jpg" />
                <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css"
                    rel="stylesheet" integrity="sha384-1BmE4kWBq78iYhFldvKuhfTAU6auU8tT94WrHftjDbrCEXSU1oBoqyl2QvZ6jIW3" crossOrigin="anonymous" />
                {/* <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.7.2/font/bootstrap-icons.css" /> */}
            </Head>
            <body>
                <Main />
                <NextScript />
            </body>
        </Html>
    )
}