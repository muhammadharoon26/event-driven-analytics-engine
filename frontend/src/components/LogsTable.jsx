import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { CheckCircle2, Clock, Loader2, ArrowRight } from 'lucide-react';

const StatusBadge = ({ status }) => {
  if (status === 'ingesting') {
    return (
      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20">
        <Loader2 className="w-3 h-3 mr-1 animate-spin" /> API
      </span>
    );
  }
  if (status === 'queued') {
    return (
      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-500/10 text-purple-400 border border-purple-500/20 shadow-[0_0_10px_rgba(168,85,247,0.2)]">
        <Clock className="w-3 h-3 mr-1 animate-pulse" /> Kafka
      </span>
    );
  }
  return (
    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
      <CheckCircle2 className="w-3 h-3 mr-1" /> Postgres
    </span>
  );
};

const LogsTable = ({ logs }) => {
  return (
    <div className="glass-panel overflow-hidden flex flex-col h-[400px]">
      <div className="p-4 border-b border-border bg-black/20 flex justify-between items-center">
        <h3 className="font-semibold flex items-center">
          <div className="w-2 h-2 rounded-full bg-emerald-500 mr-2 animate-pulse"></div>
          Live Event Stream
        </h3>
        <span className="text-xs text-gray-500 font-mono">Tracing Ingestion Flow</span>
      </div>
      
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {logs.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-gray-500 space-y-3 opacity-50">
            <Activity className="w-8 h-8" />
            <p>Waiting for events...</p>
          </div>
        ) : (
          <AnimatePresence initial={false}>
            {logs.map((log) => (
              <motion.div
                key={log.id}
                initial={{ opacity: 0, height: 0, scale: 0.95 }}
                animate={{ opacity: 1, height: 'auto', scale: 1 }}
                exit={{ opacity: 0, height: 0 }}
                transition={{ duration: 0.3 }}
                className="bg-black/40 border border-white/5 rounded-lg p-3 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-sm"
              >
                <div className="flex items-center space-x-4">
                  <div className="text-gray-400 font-mono text-xs">{log.time}</div>
                  <div>
                    <span className="font-medium text-white">{log.action}</span>
                    <span className="text-gray-500 mx-2">|</span>
                    <span className="text-gray-400">{log.user_id}</span>
                  </div>
                </div>
                
                <div className="flex items-center space-x-2">
                  <StatusBadge status="ingesting" />
                  <ArrowRight className={`w-3 h-3 ${log.status === 'queued' || log.status === 'persisted' ? 'text-gray-500' : 'text-gray-700'}`} />
                  {log.status === 'ingesting' ? (
                    <span className="w-16"></span>
                  ) : (
                    <StatusBadge status="queued" />
                  )}
                  <ArrowRight className={`w-3 h-3 ${log.status === 'persisted' ? 'text-gray-500' : 'text-gray-700'}`} />
                  {log.status === 'persisted' && <StatusBadge status="persisted" />}
                </div>
              </motion.div>
            ))}
          </AnimatePresence>
        )}
      </div>
    </div>
  );
};

// Also require Activity here for empty state
import { Activity } from 'lucide-react';

export default LogsTable;
