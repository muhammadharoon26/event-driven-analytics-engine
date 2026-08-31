import React, { useState } from 'react';
import { Zap } from 'lucide-react';
import EventSender from './components/EventSender';
import LiveStats from './components/LiveStats';
import LogsTable from './components/LogsTable';

function App() {
  const [stats, setStats] = useState({
    apiRequests: 0,
    kafkaEvents: 0,
    dbWrites: 0,
    latency: null
  });

  const [logs, setLogs] = useState([]);

  // handleEventSent is called by EventSender once the POST has actually
  // returned. `result.serverLatencyMs` is the real duration measured by the Go
  // LatencyMiddleware and read back off the X-Response-Time response header —
  // nothing here is simulated.
  const handleEventSent = (eventData, result = {}) => {
    const { ok = false, serverLatencyMs, roundTripMs, error } = result;

    const newLog = {
      id: Date.now().toString(),
      ...eventData,
      status: ok ? 'queued' : 'failed',
      error,
      serverLatencyMs,
      roundTripMs,
      time: new Date().toLocaleTimeString()
    };

    setLogs(prev => [newLog, ...prev].slice(0, 10)); // Keep last 10

    if (!ok) {
      // The API rejected it or is down. Do not advance the counters — a red
      // row is the honest outcome here.
      return;
    }

    // The API returned 202, which means the event is on the Kafka topic.
    // Round-trip is what the browser actually observed (network + handler);
    // it is the number a user would feel, so that is what the tile shows.
    setStats(prev => ({
      ...prev,
      apiRequests: prev.apiRequests + 1,
      kafkaEvents: prev.kafkaEvents + 1,
      latency: roundTripMs ?? serverLatencyMs ?? prev.latency
    }));

    // Poll for this specific event to see when the Python worker lands it in Postgres.
    pollEventStatus(newLog.id, eventData.user_id, eventData.action);
  };

  const pollEventStatus = (logId, userId, action) => {
    const maxAttempts = 10;
    let attempts = 0;
    
    const poll = setInterval(async () => {
      attempts++;
      try {
        const response = await fetch(`http://localhost:8080/api/v1/events/status/${userId}?action=${action}`);
        const data = await response.json();
        
        if (data.status === 'persisted') {
          clearInterval(poll);
          setStats(prev => ({ ...prev, dbWrites: prev.dbWrites + 1 }));
          setLogs(prev => prev.map(l => l.id === logId ? { ...l, status: 'persisted' } : l));
        } else if (attempts >= maxAttempts) {
          clearInterval(poll);
          // 10s went by and the worker never wrote it. Say so rather than
          // leaving the row stuck on "Kafka" forever.
          setLogs(prev => prev.map(l => l.id === logId ? { ...l, status: 'stalled' } : l));
        }
      } catch (e) {
        console.error("Polling error", e);
        if (attempts >= maxAttempts) {
          clearInterval(poll);
          setLogs(prev => prev.map(l => l.id === logId ? { ...l, status: 'stalled' } : l));
        }
      }
    }, 1000); // Poll every 1 second
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
