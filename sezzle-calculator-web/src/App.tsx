import Calculator from './components/Calculator'

export default function App() {
  return (
    <div className="app">
      <main className="window">
        <div className="window__titlebar">
          <span className="window__light window__light--close" aria-hidden="true" />
          <span className="window__light window__light--minimize" aria-hidden="true" />
          <span className="window__light window__light--zoom" aria-hidden="true" />
          <h1 className="window__title">Calculator</h1>
        </div>
        <Calculator />
      </main>
    </div>
  )
}
