import { StrictMode } from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import Calculator from './Calculator'
import { jsonResponse } from '../test/http'

function display(): HTMLElement {
  return screen.getByTestId('display')
}

async function pressAll(user: ReturnType<typeof userEvent.setup>, labels: string[]) {
  for (const label of labels) {
    await user.click(screen.getByRole('button', { name: label }))
  }
}

describe('Calculator', () => {
  let fetchMock: ReturnType<typeof vi.fn>

  beforeEach(() => {
    fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
  })

  it('starts at zero with the full keypad', () => {
    render(<Calculator />)
    expect(display()).toHaveTextContent('0')
    expect(screen.getByRole('button', { name: 'All clear' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Equals' })).toBeInTheDocument()
  })

  it('builds a number from digits and groups thousands', async () => {
    const user = userEvent.setup()
    render(<Calculator />)

    await pressAll(user, ['1', '2', '3', '4'])
    expect(display()).toHaveTextContent('1,234')
  })

  it('sends the operands to the API and shows the result', async () => {
    const user = userEvent.setup()
    fetchMock.mockResolvedValue(jsonResponse({ result: 42 }))
    render(<Calculator />)

    await pressAll(user, ['6', 'Multiply', '7', 'Equals'])

    expect(await screen.findByText('42')).toBeInTheDocument()
    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(String(fetchMock.mock.calls[0][0])).toContain('first=6&second=7&op=multiply')
  })

  it('evaluates the pending operation when a second operator is pressed', async () => {
    const user = userEvent.setup()
    fetchMock.mockResolvedValue(jsonResponse({ result: 5 }))
    render(<Calculator />)

    await pressAll(user, ['2', 'Add', '3', 'Add'])

    await waitFor(() => expect(display()).toHaveTextContent('5'))
    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(String(fetchMock.mock.calls[0][0])).toContain('first=2&second=3&op=sum')
    expect(screen.getByRole('button', { name: 'Add' })).toHaveAttribute('aria-pressed', 'true')
  })

  it('marks the pending operator as armed', async () => {
    const user = userEvent.setup()
    render(<Calculator />)

    await pressAll(user, ['8', 'Divide'])
    expect(screen.getByRole('button', { name: 'Divide' })).toHaveAttribute(
      'aria-pressed',
      'true',
    )
    expect(screen.getByRole('button', { name: 'Add' })).toHaveAttribute('aria-pressed', 'false')
  })

  it('negates without calling the API', async () => {
    const user = userEvent.setup()
    render(<Calculator />)

    await pressAll(user, ['5', '0', 'Negate'])

    expect(display()).toHaveTextContent('-50')
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('sends percent to the API as a unary operation', async () => {
    const user = userEvent.setup()
    fetchMock.mockResolvedValue(jsonResponse({ result: 0.5 }))
    render(<Calculator />)

    await pressAll(user, ['5', '0', 'Percent'])

    await waitFor(() => expect(display()).toHaveTextContent('0.5'))
    expect(String(fetchMock.mock.calls[0][0])).toContain('first=50&second=0&op=percentage')
  })

  it('sends square root to the API and shows it in the history', async () => {
    const user = userEvent.setup()
    fetchMock.mockResolvedValue(jsonResponse({ result: 12 }))
    render(<Calculator />)

    await pressAll(user, ['1', '4', '4', 'Square root'])

    await waitFor(() => expect(display()).toHaveTextContent('12'))
    expect(String(fetchMock.mock.calls[0][0])).toContain('first=144&second=0&op=sqrt')
    expect(screen.getByTestId('history-line')).toHaveTextContent('√144 =')
  })

  it('treats exponentiation as a binary operator', async () => {
    const user = userEvent.setup()
    fetchMock.mockResolvedValue(jsonResponse({ result: 1024 }))
    render(<Calculator />)

    await pressAll(user, ['2', 'Exponentiation'])
    expect(display()).toHaveTextContent('2 ^')
    expect(screen.getByRole('button', { name: 'Exponentiation' })).toHaveAttribute(
      'aria-pressed',
      'true',
    )

    await pressAll(user, ['1', '0', 'Equals'])

    await waitFor(() => expect(display()).toHaveTextContent('1,024'))
    expect(String(fetchMock.mock.calls[0][0])).toContain('first=2&second=10&op=exponentiation')
  })

  it('surfaces a square root error from the API', async () => {
    const user = userEvent.setup()
    fetchMock.mockResolvedValue(
      jsonResponse({ message: 'cannot take the square root of a negative number' }, 500),
    )
    render(<Calculator />)

    await pressAll(user, ['9', 'Negate', 'Square root'])

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'cannot take the square root of a negative number',
    )
    expect(display()).toHaveTextContent('Error')
  })

  it('toggles the clear key between AC and C', async () => {
    const user = userEvent.setup()
    render(<Calculator />)

    expect(screen.getByRole('button', { name: 'All clear' })).toHaveTextContent('AC')

    await pressAll(user, ['9'])
    const clearKey = screen.getByRole('button', { name: 'Clear' })
    expect(clearKey).toHaveTextContent('C')

    await user.click(clearKey)
    expect(display()).toHaveTextContent('0')
    expect(screen.getByRole('button', { name: 'All clear' })).toBeInTheDocument()
  })

  it('surfaces a server error without exposing the request id', async () => {
    const user = userEvent.setup()
    fetchMock.mockResolvedValue(
      jsonResponse({ message: 'cannot divide by zero', requestId: 'XTAqCG' }, 400),
    )
    render(<Calculator />)

    await pressAll(user, ['8', 'Divide', '0', 'Equals'])

    const alert = await screen.findByRole('alert')
    expect(alert).toHaveTextContent('cannot divide by zero')
    expect(alert).not.toHaveTextContent('XTAqCG')
    expect(alert).not.toHaveTextContent('Request ID')
    expect(display()).toHaveTextContent('Error')
  })

  it('recovers from an error with AC', async () => {
    const user = userEvent.setup()
    fetchMock.mockResolvedValue(jsonResponse({ message: 'cannot divide by zero' }, 400))
    render(<Calculator />)

    await pressAll(user, ['8', 'Divide', '0', 'Equals'])
    expect(await screen.findByRole('alert')).toBeInTheDocument()

    await pressAll(user, ['All clear'])
    expect(display()).toHaveTextContent('0')
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('reports an unreachable API', async () => {
    const user = userEvent.setup()
    fetchMock.mockRejectedValue(new TypeError('Failed to fetch'))
    render(<Calculator />)

    await pressAll(user, ['1', 'Add', '2', 'Equals'])

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Could not reach the calculator service.',
    )
  })

  it('accepts keyboard input', async () => {
    const user = userEvent.setup()
    fetchMock.mockResolvedValue(jsonResponse({ result: 9 }))
    render(<Calculator />)

    await user.keyboard('4+5{Enter}')

    await waitFor(() => expect(display()).toHaveTextContent('9'))
    expect(String(fetchMock.mock.calls[0][0])).toContain('first=4&second=5&op=sum')
  })

  it('clears with Escape and deletes with Backspace', async () => {
    const user = userEvent.setup()
    render(<Calculator />)

    await user.keyboard('123')
    expect(display()).toHaveTextContent('123')

    await user.keyboard('{Backspace}')
    expect(display()).toHaveTextContent('12')

    await user.keyboard('{Escape}')
    expect(display()).toHaveTextContent('0')
  })

  it('shows the problem being typed on the lower line', async () => {
    const user = userEvent.setup()
    fetchMock.mockResolvedValue(jsonResponse({ result: 14 }))
    render(<Calculator />)

    await pressAll(user, ['7'])
    expect(display()).toHaveTextContent('7')

    await pressAll(user, ['Multiply'])
    expect(display()).toHaveTextContent('7 ×')

    await pressAll(user, ['2'])
    expect(display()).toHaveTextContent('7 × 2')

    await pressAll(user, ['Equals'])
    await waitFor(() => expect(display()).toHaveTextContent('14'))
  })

  it('leaves the upper line empty until a calculation completes', async () => {
    const user = userEvent.setup()
    fetchMock.mockResolvedValue(jsonResponse({ result: 14 }))
    render(<Calculator />)

    const upper = () => screen.getByTestId('history-line')

    expect(upper()).toBeEmptyDOMElement()

    await pressAll(user, ['7', 'Multiply', '2'])
    expect(upper()).toBeEmptyDOMElement()

    await pressAll(user, ['Equals'])
    await waitFor(() => expect(upper()).toHaveTextContent('7 × 2 ='))
    expect(display()).toHaveTextContent('14')
  })

  it('keeps the previous problem on the upper line while the next is typed', async () => {
    const user = userEvent.setup()
    fetchMock.mockResolvedValueOnce(jsonResponse({ result: 14 }))
    render(<Calculator />)

    await pressAll(user, ['7', 'Multiply', '2', 'Equals'])
    await waitFor(() => expect(screen.getByTestId('history-line')).toHaveTextContent('7 × 2 ='))

    await pressAll(user, ['9', 'Add'])
    expect(screen.getByTestId('history-line')).toHaveTextContent('7 × 2 =')
    expect(display()).toHaveTextContent('9 +')
  })

  it('cancels an in-flight request when clear is pressed', async () => {
    const user = userEvent.setup()
    fetchMock.mockReturnValue(new Promise<Response>(() => {}))
    render(<Calculator />)

    await pressAll(user, ['1', 'Add', '2', 'Equals'])
    await waitFor(() => expect(screen.getByRole('button', { name: '7' })).toBeDisabled())

    await pressAll(user, ['Clear'])

    expect(display()).toHaveTextContent('1 +')
    expect(screen.getByRole('button', { name: '7' })).toBeEnabled()

    await pressAll(user, ['All clear'])
    expect(display()).toHaveTextContent('0')
  })

  it('fires only one request when two keys land in the same tick', async () => {
    fetchMock.mockResolvedValue(jsonResponse({ result: 3 }))
    render(<Calculator />)

    const press = (key: string) =>
      fireEvent.keyDown(window, { key })

    press('1')
    press('+')
    press('2')
    press('Enter')
    press('Enter')

    await waitFor(() => expect(display()).toHaveTextContent('3'))
    expect(fetchMock).toHaveBeenCalledTimes(1)
  })

  it('applies a result exactly once under StrictMode', async () => {
    const user = userEvent.setup()
    fetchMock.mockResolvedValue(jsonResponse({ result: 3 }))
    render(
      <StrictMode>
        <Calculator />
      </StrictMode>,
    )

    await pressAll(user, ['1', 'Add', '2', 'Equals'])

    await waitFor(() => expect(display()).toHaveTextContent('3'))
    // StrictMode runs effects twice; the request must still be sent once.
    expect(fetchMock).toHaveBeenCalledTimes(1)
  })

  it('disables the keypad while a request is in flight', async () => {
    const user = userEvent.setup()
    let release: (value: Response) => void = () => {}
    fetchMock.mockReturnValue(
      new Promise<Response>((resolve) => {
        release = resolve
      }),
    )
    render(<Calculator />)

    await pressAll(user, ['1', 'Add', '2', 'Equals'])

    await waitFor(() =>
      expect(screen.getByRole('button', { name: '7' })).toBeDisabled(),
    )

    release(jsonResponse({ result: 3 }))
    await waitFor(() => expect(display()).toHaveTextContent('3'))
    expect(screen.getByRole('button', { name: '7' })).toBeEnabled()
  })
})
