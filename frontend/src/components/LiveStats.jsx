import React from 'react';
import { motion } from 'framer-motion';
import { Activity, Database, Server, Clock } from 'lucide-react';

const StatCard = ({ title, value, icon: Icon, colorClass, delay }) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    animate={{ opacity: 1, y: 0 }}
    transition={{ delay, duration: 0.5 }}
    className="glass-panel p-5 relative overflow-hidden group"
  >
    <div className={`absolute -right-6 -top-6 w-24 h-24 rounded-full blur-2xl opacity-20 group-hover:opacity-40 transition-opacity ${colorClass.replace('text-', 'bg-')}`}></div>
    
    <div className="flex justify-between items-start relative z-10">
      <div>
        <p className="text-sm font-medium text-gray-400 mb-1">{title}</p>
        <div className="flex items-baseline space-x-2">
          <h3 className="text-3xl font-bold tracking-tight">{value}</h3>
          {(title === 'Avg Latency') && <span className="text-sm text-gray-500">ms</span>}
        </div>
      </div>
      <div className={`p-3 rounded-xl bg-white/5 border border-white/5 ${colorClass}`}>
        <Icon className="w-5 h-5" />
      </div>
    </div>
  </motion.div>
);

const LiveStats = ({ stats }) => {
  return (
    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
      <StatCard 
        title="API Accepted" 
        value={stats.apiRequests} 
        icon={Activity} 
        colorClass="text-blue-400" 
        delay={0.1}
      />
      <StatCard 
        title="Kafka Enqueued" 
        value={stats.kafkaEvents} 
        icon={Server} 
        colorClass="text-purple-400" 
        delay={0.2}
      />
      <StatCard 
        title="DB Written" 
        value={stats.dbWrites} 
        icon={Database} 
        colorClass="text-emerald-400" 
        delay={0.3}
      />
      <StatCard 
        title="Avg Latency" 
        value={stats.latency} 
        icon={Clock} 
        colorClass="text-amber-400" 
        delay={0.4}
      />
    </div>
  );
};

export default LiveStats;
