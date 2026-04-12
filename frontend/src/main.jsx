import React from 'react'
import { createRoot } from 'react-dom/client'
import { App as AntdApp } from 'antd'
import 'antd/dist/reset.css'
import './style.css'
import App from './App'

const container = document.getElementById('root')

const root = createRoot(container)

root.render(
  <React.StrictMode>
    <AntdApp>
      <App />
    </AntdApp>
  </React.StrictMode>
)
