import type { NextPage } from 'next'
import Head from 'next/head'
import Image from 'next/image'
import { FormEvent, useState } from 'react'
import { Alert } from 'react-bootstrap'

const Home: NextPage = () => {

  type AlertState = {
    message?: string
    variant?: 'success' | 'warning' | 'danger'
  }

  const [email, setEmail] = useState('')
  const [alertState, setAlertState] = useState<AlertState>({})

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault()
    let response = await fetch('/api/404', { // /send-invite
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
  }

  function isEmailValid(value: string): boolean {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)
  }

  return (
    <div className="container w-sm-50">
      <Head>
        <title>hello there</title>
        <meta name="description" content="invite yourself to my plex server" />
        <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no"></meta>
        <link rel="icon" href="https://avatars.slack-edge.com/2022-01-11/2950060844657_4cae9e95e482718f4ef6_88.jpg" />
      </Head>

      <main>
        <div className="row justify-content-center">
          <div className="col-12 p-3 text-center">
            <h1>hello there</h1>
            invite yourself to my plex server
          </div>

          <div className='col-sm-6 p-3' >
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
                  <button id='button' type='submit'
                    className='btn btn-primary'
                    disabled={!isEmailValid(email)}>send invite</button>
                </div>
              </div>
            </form>
          </div>

          <div className='w-100' />


          <div className='col-sm-6 p-3'>
            <a
              href="https://join.slack.com/t/bagziga/shared_invite/zt-a2uj179c-hvwdWXLf3g0mT1eNqAG_KQ"
              target="_blank"
              rel="noopener noreferrer"
              className="align-middle"
            >
              join the
              <Image src='/slack.png' alt='Slack logo' height={30} width={73} />
              and post any issues you might have
            </a>
          </div>

        </div>

      </main>
    </div>
  )
}

export default Home

