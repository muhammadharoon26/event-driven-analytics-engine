import React, { useState } from 'react';
import { Send, Smartphone, ShoppingCart, Eye, MousePointerClick, AlertTriangle } from 'lucide-react';
import axios from 'axios';

const ACTIONS = [
  { id: 'scan_product', label: 'Scan Product', icon: Smartphone, color: 'text-blue-400' },
  { id: 'add_to_cart', label: 'Add to Cart', icon: ShoppingCart, color: 'text-emerald-400' },
  { id: 'page_view', label: 'Page View', icon: Eye, color: 'text-purple-400' },
  { id: 'button_click', label: 'Button Click', icon: MousePointerClick, color: 'text-amber-400' },
];

// Go's time.Duration.String() emits things like "512.7µs", "1.204ms", "1.05s".
// Convert whatever the API sent into plain milliseconds.
const parseGoDuration = (raw) => {
  if (!raw) return null;
  const match = String(raw).trim().match(/^([\d.]+)\s*(ns|µs|us|ms|s)$/);
  if (!match) return null;
  const value = parseFloat(match[1]);
  if (Number.isNaN(value)) return null;
  const toMs = { ns: 1e-6, 'µs': 1e-3, us: 1e-3, ms: 1, s: 1000 };
  return value * toMs[match[2]];
};

const EventSender = ({ onSend }) => {
  const [selectedAction, setSelectedAction] = useState(ACTIONS[0].id);
  const [userId, setUserId] = useState('user-' + Math.floor(Math.random() * 10000));
  const [isSending, setIsSending] = useState(false);
  const [error, setError] = useState(null);

  const handleSend = async () => {
    setIsSending(true);
    setError(null);
    const eventPayload = {
      user_id: userId,
      action: selectedAction,
      timestamp: new Date().toISOString()
    };

    const startedAt = performance.now();

    try {
      // In a real environment, this points to your deployed Go API
      const response = await axios.post('http://localhost:8080/api/v1/events', eventPayload);
      const roundTripMs = performance.now() - startedAt;

      // X-Response-Time is set by LatencyMiddleware in ingestion-api/telemetry/middleware.go
      // and exposed to the browser via Access-Control-Expose-Headers in main.go.
      const serverMs = parseGoDuration(response.headers['x-response-time']);

      onSend(eventPayload, {
        ok: true,
        serverLatencyMs: serverMs,
        roundTripMs
      });

      // Only roll a fresh user ID once the event actually made it in, so a
      // retry after a failure re-sends the same identity.
      setUserId('user-' + Math.floor(Math.random() * 10000));
    } catch (err) {
      console.error("Failed to send event:", err);
      const message = err.response
        ? `API returned ${err.response.status}`
        : 'API unreachable on :8080';
      setError(message);
      onSend(eventPayload, { ok: false, error: message });
    } finally {
      setIsSending(false);
    }
  };

  return (
    <div className="glass-panel p-6">
      <div className="flex items-center space-x-3 mb-6">
        <Send className="w-5 h-5 text-primary" />
        <h2 className="text-xl font-semibold">Emit Event</h2>
      </div>

      <div className="space-y-5">
        <div>
          <label className="block text-sm text-gray-400 mb-2">User ID</label>
          <input 
            type="text" 
            value={userId}
            onChange={(e) => setUserId(e.target.value)}
            className="glass-input w-full"
          />
        </div>

        <div>
          <label className="block text-sm text-gray-400 mb-2">Action Type</label>
          <div className="grid grid-cols-2 gap-3">
            {ACTIONS.map((action) => {
              const Icon = action.icon;
              const isSelected = selectedAction === action.id;
              return (
                <button
                  key={action.id}
                  onClick={() => setSelectedAction(action.id)}
                  className={`flex flex-col items-center justify-center p-4 rounded-xl border transition-all duration-200 ${
                    isSelected 
                      ? 'border-primary bg-primary/10 scale-100 shadow-[0_0_15px_rgba(59,130,246,0.2)]' 
                      : 'border-white/10 bg-black/20 hover:bg-white/5 active:scale-95'
                  }`}
                >
                  <Icon className={`w-6 h-6 mb-2 ${action.color}`} />
                  <span className="text-xs font-medium text-gray-300">{action.label}</span>
                </button>
              );
            })}
          </div>
        </div>

        <button 
          onClick={handleSend}
          disabled={isSending}
          className="w-full py-3 mt-4 glass-button bg-primary/20 hover:bg-primary/30 border-primary/50 text-white flex items-center justify-center font-semibold text-lg overflow-hidden relative group"
        >
          <span className="relative z-10 flex items-center">
            {isSending ? (
              <span className="animate-spin mr-2 border-2 border-white border-t-transparent rounded-full w-5 h-5"></span>
            ) : (
              <Send className="w-5 h-5 mr-2 group-hover:translate-x-1 transition-transform" />
            )}
            Fire Event Asynchronously
          </span>
          {/* Shine effect */}
          <div className="absolute inset-0 -translate-x-full group-hover:animate-[shimmer_1.5s_infinite] bg-gradient-to-r from-transparent via-white/10 to-transparent z-0"></div>
        </button>

        {error && (
          <div className="flex items-center gap-2 px-3 py-2 rounded-lg border border-red-500/30 bg-red-500/10 text-red-300 text-sm">
            <AlertTriangle className="w-4 h-4 shrink-0" />
            <span>{error}</span>
          </div>
        )}
      </div>
    </div>
  );
};

export default EventSender;
