//import { useState } from 'react'
import { BrowserRouter as Router, Routes, Route, /*Navigate*/ } from 'react-router-dom';
import './App.css'
import MainPage from './components/Pages/MainPage';
import Layout from './components/common/Layout';
import TeamInitialize from './components/Pages/TeamInitialize';
import RegisterFormLayout from './components/common/RegisterFormLayout';
import LoginPage from './components/Pages/login';
import TournamentsPage from './components/Pages/TournamentsPage';
import AdminLayout from './components/Admin/AdminLayout';
import AdminPanel from './components/Admin/AdminPanel';
import TeamApplicationsPage from './components/Admin/TeamsApplications';
import Tournaments from './components/Admin/Tournaments';

function App() {
 
  return (
    <Router>
       <div className="App">
      <Routes>
          <Route element={<Layout />}>
          <Route path="/" element={<MainPage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/tournaments" element={<TournamentsPage />} />
          </Route>
           <Route element={<AdminLayout />}>
            <Route path="/AdminPanel" element={<AdminPanel />} />
            <Route path='/TeamsApplications' element={<TeamApplicationsPage/>} />
            <Route path='/olymp-admin/event/index' element={<Tournaments/>}/>
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
