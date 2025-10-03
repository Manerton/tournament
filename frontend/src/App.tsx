import { useState } from 'react'
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import './App.css'
import RegisterTeamPage from './components/Pages/TeamInitialize';

function App() {
 
  return (
    <Router>
       <div className="App">
      <Routes>
        <Route path="/" element={<RegisterTeamPage />} />
      </Routes>
    </div>
    </Router>
  )
}

export default App
