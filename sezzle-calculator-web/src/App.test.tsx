import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import App from './App'

describe('App', () => {
  it('renders the calculator window', () => {
    render(<App />)
    expect(screen.getByRole('heading', { name: 'Calculator' })).toBeInTheDocument()
    expect(screen.getByTestId('display')).toHaveTextContent('0')
    expect(screen.getByRole('button', { name: 'Equals' })).toBeInTheDocument()
  })
})
