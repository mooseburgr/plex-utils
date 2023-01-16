import type { NextPage } from 'next'
import Head from 'next/head'
import { FormEvent, useState } from 'react'
import { Alert, Button, Card } from 'react-bootstrap'
import SlackIcon from './slack-icon'

const Home: NextPage = () => {

  type AlertState = {
    message?: string
    variant?: 'success' | 'warning' | 'danger'
  }

  const [email, setEmail] = useState('')
  const [isLoading, setLoading] = useState(false);
  const [alertState, setAlertState] = useState<AlertState>({})

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault()
    setLoading(true)

    await fetch('https://send-plex-invite-d3zjrv6fua-uc.a.run.app', {
      method: 'POST',
      body: JSON.stringify({ email: email }),
    })
      .then(async response => {
        if (response.ok) {
          setAlertState({ message: 'invite sent! check your email', variant: 'success' })
        } else if (response.status === 422) {
          // email is already in use or pending
          setAlertState({ message: 'looks like that email is already in use or pending, check your email', variant: 'warning' })
        } else {
          throw new Error(await response.text())
        }
      })
      .catch(err => {
        console.error(err)
        setAlertState({ message: 'oops...something went wrong: ' + err, variant: 'danger' })
      })

    setLoading(false)
  }

  function isEmailValid(value: string): boolean {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)
  }

  return (
    <div className="container">
      <Head>
        <title>hello there</title>
        <meta name="description" content="invite yourself to my plex server" />
        <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no"></meta>
        <link rel="icon" href="https://avatars.slack-edge.com/2022-01-11/2950060844657_4cae9e95e482718f4ef6_88.jpg" />
        <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css"
          rel="stylesheet" integrity="sha384-1BmE4kWBq78iYhFldvKuhfTAU6auU8tT94WrHftjDbrCEXSU1oBoqyl2QvZ6jIW3" crossOrigin="anonymous" />
        {/* <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.7.2/font/bootstrap-icons.css" /> */}
      </Head>

      <main>
        <div className="row justify-content-center">
          <div className="col-12 p-3 text-center">
            <h1 onClick={() => window.open('https://www.youtube.com/watch?v=rEq1Z0bjdwc', '_blank')}>
              hello there
            </h1>
            invite yourself to my plex server

            <a rel="me" href="https://hachyderm.io/@mooseburger" className="d-none">hi mastodon</a>
          </div>

          <div className='col-sm-6 p-3' >
            <Card className='shadow'>
              <Card.Body>
            {alertState.message &&
              <Alert variant={alertState.variant} dismissible onClose={() => setAlertState({})}>
                {alertState.message}
              </Alert>
            }

            <form onSubmit={handleSubmit}>
              <div className='input-group'>
                <label htmlFor='email' className='d-none'>your email</label>
                <input
                  id='email'
                  type="text"
                  name="email"
                  placeholder='your email'
                  onChange={event => setEmail(event.target.value)}
                  required
                  aria-label="Email"
                  className='form-control'
                />
                <div className="input-group-append">
                  <Button
                    disabled={!isEmailValid(email) || isLoading}
                    type='submit'
                  >
                    {isLoading ? 'loading...' : 'send invite'}
                  </Button>
                </div>
              </div>
            </form>
              </Card.Body>
            </Card>
          </div>

          <div className='w-100' />

          <div className='col-sm-3 p-3'>
            <Card className='shadow'>
              <Card.Body>
                <Card.Text>
                  hit up the #plex channel for updates, content requests, or any issues you might have
                </Card.Text>
                <Card.Link href="https://join.slack.com/t/bagziga/shared_invite/zt-a2uj179c-hvwdWXLf3g0mT1eNqAG_KQ" target="_blank" rel="noopener noreferrer">
                  <SlackIcon /> <span className='p-1'>join the slack</span>
                </Card.Link>
              </Card.Body>
            </Card>
          </div>

          <div className='col-sm-3 p-3'>
            <Card className='shadow'>
              <Card.Body>
                <Card.Text>
                  here&apos;s how to tweak the default settings to optimize streaming quality
                </Card.Text>
                <Card.Link href="https://www.aaviah.com/plex" target="_blank" rel="noopener noreferrer">
                  h/t @aaviah
                </Card.Link>
              </Card.Body>
            </Card>
          </div>

        </div>

      </main>
    </div>
  )
}

export default Home

