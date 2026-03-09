import React, { useState } from 'react';
import { Send, Smartphone, ShoppingCart, Eye, MousePointerClick } from 'lucide-react';
import axios from 'axios';

const ACTIONS = [
  { id: 'scan_product', label: 'Scan Product', icon: Smartphone, color: 'text-blue-400' },
  { id: 'add_to_cart', label: 'Add to Cart', icon: ShoppingCart, color: 'text-emerald-400' },
  { id: 'page_view', label: 'Page View', icon: Eye, color: 'text-purple-400' },
  { id: 'button_click', label: 'Button Click', icon: MousePointerClick, color: 'text-amber-400' },
];

const EventSender = ({ onSend }) => {
  const [selectedAction, setSelectedAction] = useState(ACTIONS[0].id);
  const [userId, setUserId] = useState('user-' + Math.floor(Math.random() * 10000));
  const [isSending, setIsSending] = useState(false);

  const handleSend = async () => {
    setIsSending(true);
    const eventPayload = {
      user_id: userId,
      action: selectedAction,
      timestamp: new Date().toISOString()
    };

    try {
      // In a real environment, this points to your deployed Go API
      await axios.post('http://localhost:8080/api/v1/events', eventPayload);
      onSend(eventPayload);
    } catch (error) {
      console.error("Failed to send event:", error);
      // Still trigger UI update for demonstration if API is down
      onSend(eventPayload);
    } finally {
      setIsSending(false);
      // Generate new user ID for next event
      setUserId('user-' + Math.floor(Math.random() * 10000));
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
      </div>
    </div>
  );
};

export default EventSender;
