import { useState } from 'react'
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import './App.css'
import RegisterTeamPage from './components/Pages/TeamInitialize';
import MainPage from './components/Pages/MainPage';
import Layout from './components/common/Layout';
import TeamInitialize from './components/Pages/TeamInitialize';
import RegisterFormLayout from './components/common/RegisterFormLayout';

function App() {
 
  return (
    <Router>
       <div className="App">
      <Routes>
          <Route element={<Layout />}>
          <Route path="/" element={<MainPage />} />
          
          </Route>
          <Route element={<RegisterFormLayout />}>
            <Route path="/tournament/registerTeam/:id" element={<TeamInitialize/>} />
          </Route>
      </Routes>
    </div>
    </Router>
  )
}

export default App
