import { useCallback, useEffect, useLayoutEffect, useReducer, useRef, useState } from 'react'
import { calculate } from '../api/client'
import { toDisplayError } from '../api/ApiError'
import { OPERATIONS } from '../api/types'
import {
  armedOperation,
  clearAriaLabel,
  clearLabel,
  fitFontSize,
  historyLine,
  INITIAL_STATE,
  inputLine,
  interrupts,
  MIN_HISTORY_FONT_PX,
  MIN_VALUE_FONT_PX,
  reduce,
} from '../lib/calculator'
import { KEY_BY_CHARACTER, PAD_KEYS } from '../lib/keypad'

const FIT_SAFETY_PX = 2

function useFitText(text: string, minPx: number) {
  const boxRef = useRef<HTMLDivElement | null>(null)
  const textRef = useRef<HTMLSpanElement | null>(null)
  const [clipped, setClipped] = useState(false)

  const fit = useCallback(() => {
    const box = boxRef.current
    const el = textRef.current
    if (!box || !el) return

    el.style.fontSize = ''
    const base = Number.parseFloat(getComputedStyle(el).fontSize)
    const available = box.clientWidth - FIT_SAFETY_PX
    let width = el.getBoundingClientRect().width

    if (!base || available <= 0 || !width) {
      setClipped(false)
      return
    }

    let size = base
    for (let pass = 0; pass < 2; pass += 1) {
      const next = fitFontSize(size, minPx, width, available)
      if (next >= size) break
      size = next
      el.style.fontSize = `${size}px`
      width = el.getBoundingClientRect().width
    }

    setClipped(width > available + 0.5)
  }, [minPx])

  useLayoutEffect(() => {
    fit()
  }, [fit, text])

  useEffect(() => {
    const box = boxRef.current
    if (!box) return

    let frame = 0
    const observer = new ResizeObserver(() => {
      cancelAnimationFrame(frame)
      frame = requestAnimationFrame(fit)
    })
    observer.observe(box)

    return () => {
      cancelAnimationFrame(frame)
      observer.disconnect()
    }
  }, [fit])

  return { boxRef, textRef, clipped }
}

export default function Calculator() {
  const [state, dispatch] = useReducer(reduce, INITIAL_STATE)
  const [activeKey, setActiveKey] = useState<string | null>(null)

  const flashRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const pending = state.inFlight !== null

  useEffect(() => {
    const request = state.inFlight
    if (!request) return

    const controller = new AbortController()
    let live = true

    void calculate(request, { signal: controller.signal }).then(
      ({ result }) => {
        if (!live) return
        dispatch({ type: 'result', result })
      },
      (error: unknown) => {
        if (!live || controller.signal.aborted) return
        dispatch({ type: 'failure', message: toDisplayError(error) })
      },
    )

    return () => {
      live = false
      controller.abort()
    }
  }, [state.inFlight])

  useEffect(
    () => () => {
      if (flashRef.current) clearTimeout(flashRef.current)
    },
    [],
  )

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (event.metaKey || event.ctrlKey || event.altKey) return
      const key = KEY_BY_CHARACTER[event.key]
      if (!key) return

      event.preventDefault()
      dispatch(key.action)

      setActiveKey(key.id)
      if (flashRef.current) clearTimeout(flashRef.current)
      flashRef.current = setTimeout(() => setActiveKey(null), 120)
    }

    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [])

  const valueText = inputLine(state)
  const historyText = historyLine(state)
  const valueFit = useFitText(valueText, MIN_VALUE_FONT_PX)
  const historyFit = useFitText(historyText, MIN_HISTORY_FONT_PX)

  const armed = armedOperation(state)

  return (
    <div className="calculator" data-pending={pending || undefined}>
      <div
        className="calculator__display"
        role="status"
        aria-live="polite"
        aria-label="Display"
      >
        <div
          className="calculator__line calculator__line--history"
          ref={historyFit.boxRef}
          data-clipped={historyFit.clipped || undefined}
        >
          <span
            className="calculator__history-line"
            data-testid="history-line"
            ref={historyFit.textRef}
          >
            {historyText}
          </span>
        </div>
        <div
          className="calculator__line"
          ref={valueFit.boxRef}
          data-clipped={valueFit.clipped || undefined}
        >
          <span className="calculator__value" data-testid="display" ref={valueFit.textRef}>
            {valueText}
          </span>
        </div>
      </div>

      <div className="calculator__keys">
        {PAD_KEYS.map((key) => {
          const isClear = key.action.type === 'clear'
          const isBinary = key.action.type === 'operation' && OPERATIONS[key.action.op].arity === 2
          const isArmed = key.action.type === 'operation' && armed === key.action.op

          return (
            <button
              key={key.id}
              type="button"
              className="key"
              data-kind={key.kind}
              data-wide={key.wide || undefined}
              data-armed={isArmed || undefined}
              data-active={activeKey === key.id || undefined}
              aria-label={isClear ? clearAriaLabel(state) : key.ariaLabel}
              aria-pressed={isBinary ? isArmed : undefined}
              disabled={pending && !interrupts(key.action)}
              onClick={() => dispatch(key.action)}
            >
              {isClear ? clearLabel(state) : key.label}
            </button>
          )
        })}
      </div>

      {state.error ? (
        <p className="calculator__error" role="alert">
          {state.error}
        </p>
      ) : null}
    </div>

  )
}
