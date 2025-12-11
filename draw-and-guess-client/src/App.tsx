import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Home from './pages/Home';
import RoomList from './pages/RoomList';
import GameRoom from './pages/GameRoom';
import './App.css';

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/rooms" element={<RoomList />} />
        <Route path="/game/:roomId" element={<GameRoom />} />
        <Route path="/game/:roomId/guess" element={<GameRoom />} />
      </Routes>
    </BrowserRouter>
  );
};

export default App;
