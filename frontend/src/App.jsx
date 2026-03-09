import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { Activity, Database, Server, Zap } from 'lucide-react';
import EventSender from './components/EventSender';
import LiveStats from './components/LiveStats';
import LogsTable from './components/LogsTable';

function App() {
  const [stats, setStats] = useState({
    apiRequests: 0,
    kafkaEvents: 0,
    dbWrites: 0,
    latency: 12
  });

  const [logs, setLogs] = useState([]);

  // Simulate WebSocket / periodic polling for metrics
  useEffect(() => {
    const interval = setInterval(() => {
      // In a real app we'd fetch this from the Go/Python services
      setStats(prev => ({
        ...prev,
        latency: Math.floor(Math.random() * 20) + 5
      }));
    }, 2000);
    return () => clearInterval(interval);
  }, []);

  const handleEventSent = (eventData) => {
    // Optimistic UI updates to fake the latency of the async flow visually
    setStats(prev => ({ ...prev, apiRequests: prev.apiRequests + 1 }));
    
    // Simulate API accepting -> Kafka queuing -> Python writing
    const newLog = {
      id: Date.now().toString(),
      ...eventData,
      status: 'ingesting',
      time: new Date().toLocaleTimeString()
    };
    
    setLogs(prev => [newLog, ...prev].slice(0, 10)); // Keep last 10
    
    setTimeout(() => {
      setStats(prev => ({ ...prev, kafkaEvents: prev.kafkaEvents + 1 }));
      setLogs(prev => prev.map(l => l.id === newLog.id ? { ...l, status: 'queued' } : l));
      
      setTimeout(() => {
        setStats(prev => ({ ...prev, dbWrites: prev.dbWrites + 1 }));
        setLogs(prev => prev.map(l => l.id === newLog.id ? { ...l, status: 'persisted' } : l));
      }, 800);
    }, 300);
  };

  return (
    <div className="max-w-7xl mx-auto px-4 py-12">
      {/* Header */}
      <header className="mb-12 text-center animate-fade-in">
        <div className="inline-flex items-center justify-center p-3 glass-panel mb-6 rounded-full">
          <Zap className="w-8 h-8 text-primary" />
        </div>
        <h1 className="text-4xl md:text-5xl font-bold mb-4 tracking-tight">
          Event-Driven <span className="gradient-text">Analytics Engine</span>
        </h1>
        <p className="text-gray-400 text-lg max-w-2xl mx-auto">
          High-throughput, decoupled microservices demonstrating real-time ingestion, Kafka queuing, and persistent storage.
        </p>
      </header>

      {/* Main Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 mb-8">
        {/* Left Column - Sender */}
        <div className="lg:col-span-1 space-y-8 animate-slide-up" style={{ animationDelay: '0.1s' }}>
          <EventSender onSend={handleEventSent} />
        </div>

        {/* Right Column - Stats & Flow */}
        <div className="lg:col-span-2 space-y-8 animate-slide-up" style={{ animationDelay: '0.2s' }}>
          <LiveStats stats={stats} />
          <LogsTable logs={logs} />
        </div>
      </div>
    </div>
  );
}

export default App;
