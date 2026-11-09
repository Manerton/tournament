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
import ToursPage from './components/Pages/ToursPage';
import GamesPage from './components/Pages/Games';
import GameVotePage from './components/Pages/GameVote';
import PlayersTemplate from "./components/Leaderboards/PlayersTemplate.tsx";
import ScoreTemplate from "./components/Leaderboards/ScoreTemplate.tsx";
import WinLeaderboard from "./components/Leaderboards/WinLeaderboard.tsx";
import GlobalLeaderboard from "./components/Leaderboards/GlobalLeaderboard.tsx";
import UsersManagement from "./components/Admin/UsersManagment.tsx";
import ReportPage from './components/Admin/Report.tsx';
import WinLeaderboardHide from "./components/Leaderboards/WinLeaderboardHide.tsx";
import QuizDisplay from "./components/Quiz/QuizDisplay.tsx";
import QuizHerald from "./components/Quiz/QuizHerald.tsx";
import QuizAdmin from "./components/Quiz/QuizAdmin.tsx";
import QuizQuestionBank from "./components/Quiz/QuizQuestionBank.tsx";


function App() {
 
  return (
    <Router>
       <div className="App">
      <Routes>
          <Route element={<Layout />}>
              <Route path="/" element={<MainPage />} />
              <Route path="/login" element={<LoginPage />} />
              <Route path="/tournaments" element={<TournamentsPage />} />
              <Route path="/tournaments/:id/tours" element={<ToursPage/>} />
              <Route path="/tours/:id/games" element={<GamesPage/>} />
              <Route path="/vote" element={<GameVotePage/>} />
              <Route path="/tournaments/:id/players" element={<PlayersTemplate />} />
              <Route path="/tournaments/:id/score" element={<ScoreTemplate />} />
              <Route path="/tournaments/:id/wins" element={<WinLeaderboard />} />
              <Route path="/tournaments/:id/winsspecial" element={<WinLeaderboardHide />} />
              <Route path="/tournaments/:id/leaderboard" element={<GlobalLeaderboard />} />
          </Route>

          {/* Quiz screens run standalone: no site header/footer or nav buttons */}
          <Route path="/quiz/display/:eventId" element={<QuizDisplay />} />
          <Route path="/quiz/herald/:eventId" element={<QuizHerald />} />

           <Route element={<AdminLayout />}>
            <Route path="/AdminPanel" element={<AdminPanel />} />
            <Route path="/quiz/admin/:eventId" element={<QuizAdmin />} />
            <Route path="/quiz/admin/:eventId/questions" element={<QuizQuestionBank />} />
            <Route path="/Report" element={<ReportPage />} />
            R
            <Route path='/TeamsApplications' element={<TeamApplicationsPage/>} />
               <Route path='/UsersManagment' element={<UsersManagement/>} />
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
