import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Home from './pages/Home';
import GameModeSelection from './pages/GameModeSelection';
import CreateRoom from './pages/CreateRoom';
import RoomList from './pages/RoomList';
import GameRoom from './pages/GameRoom';
import QuizRoom from './pages/QuizRoom';
import QuizGame from './pages/QuizGame';
import './App.css';

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/game-mode" element={<GameModeSelection />} />
        <Route path="/create-room" element={<CreateRoom />} />
        <Route path="/rooms" element={<RoomList />} />
        <Route path="/game/:roomId" element={<GameRoom />} />
        <Route path="/quiz-room/:roomId" element={<QuizRoom />} />
        <Route path="/game/:roomId/guess" element={<GameRoom />} />
        <Route path="/quiz/:type" element={<QuizGame />} />
      </Routes>
    </BrowserRouter>
  );
};

export default App;
