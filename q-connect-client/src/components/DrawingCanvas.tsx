import React, { useRef, useEffect, useState, useCallback } from 'react';
import { Stroke, Point, DrawingEventPayload } from '../types';
import { wsService } from '../services/websocket';

interface DrawingCanvasProps {
  uuid: string;
  allowDrawing: boolean;
  selectedColor: string;
  username?: string; // Add username for player identification
  onStrokesChange?: (strokes: Stroke[]) => void;
  onClearCanvas?: () => void;
  clearTrigger?: number;
}

const DrawingCanvas: React.FC<DrawingCanvasProps> = ({
  uuid,
  allowDrawing,
  selectedColor,
  username = 'Guest',
  onStrokesChange,
  clearTrigger,
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [isDrawing, setIsDrawing] = useState(false);
  const [strokes, setStrokes] = useState<Stroke[]>([]);
  const [currentStroke, setCurrentStroke] = useState<Point[]>([]);
  const [remoteStroke, setRemoteStroke] = useState<Stroke | null>(null);
  const currentStrokeId = useRef<string>('');

  const canvasWidth = 350;
  const canvasHeight = 350;

  const generateStrokeId = (): string => {
    return `stroke-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  };

  const clearCanvas = useCallback(() => {
    setStrokes([]);
    setCurrentStroke([]);
    setRemoteStroke(null);
    if (onStrokesChange) onStrokesChange([]);
  }, [onStrokesChange]);

  // Clear canvas when clearTrigger changes (new round)
  useEffect(() => {
    if (clearTrigger !== undefined && clearTrigger > 0) {
      clearCanvas();
    }
  }, [clearTrigger, clearCanvas]);

  // Subscribe to draw events (aligned with DrawingEventPayload)
  useEffect(() => {
    // Handle DRAW_EVENT messages in WSSuccessMessage format
    const drawEventHandler = (payload: DrawingEventPayload) => {
      try {
        // Ignore our own drawing events
        if (payload.playerId === username) return;

        if (payload.action === 'start') {
          setRemoteStroke({
            points: payload.points || [],
            color: payload.color || '#000000',
          });
        } else if (payload.action === 'draw' && payload.points) {
          setRemoteStroke((prev) => {
            if (!prev) return null;
            return {
              ...prev,
              points: [...prev.points, ...payload.points],
            };
          });
        } else if (payload.action === 'end') {
          setRemoteStroke((prev) => {
            if (prev && prev.points.length > 0) {
              setStrokes((prevStrokes) => [...prevStrokes, prev]);
            }
            return null;
          });
        } else if (payload.action === 'clear') {
          clearCanvas();
        }
      } catch (error) {
        console.error('Error handling draw event:', error);
      }
    };

    // Register handler for DRAW_EVENT messages
    wsService.addMessageHandler('draw_event', drawEventHandler);

    // Also subscribe to legacy draw channel for backward compatibility
    const subscription = wsService.subscribe(`draw/${uuid}`, (data) => {
      try {
        const { type, color, point } = data;

        if (type === 'start') {
          setRemoteStroke({
            points: [{ x: point.x, y: point.y }],
            color: color,
          });
        } else if (type === 'draw') {
          setRemoteStroke((prev) => {
            if (!prev) return null;
            return {
              ...prev,
              points: [...prev.points, { x: point.x, y: point.y }],
            };
          });
        } else if (type === 'end') {
          setRemoteStroke((prev) => {
            if (prev) {
              setStrokes((prevStrokes) => [...prevStrokes, prev]);
            }
            return null;
          });
        }
      } catch (error) {
        console.error('Error handling draw message:', error);
      }
    });

    return () => {
      wsService.removeMessageHandler('draw_event', drawEventHandler);
      if (subscription) {
        subscription.unsubscribe();
      }
    };
  }, [uuid, username, clearCanvas]);

  // Render canvas
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Clear canvas
    ctx.fillStyle = 'white';
    ctx.fillRect(0, 0, canvasWidth, canvasHeight);

    // Draw completed strokes
    strokes.forEach((stroke) => {
      drawStroke(ctx, stroke);
    });

    // Draw current stroke
    if (currentStroke.length > 0) {
      drawStroke(ctx, { points: currentStroke, color: selectedColor });
    }

    // Draw remote stroke
    if (remoteStroke) {
      drawStroke(ctx, remoteStroke);
    }
  }, [strokes, currentStroke, remoteStroke, selectedColor]);

  const drawStroke = (ctx: CanvasRenderingContext2D, stroke: Stroke) => {
    if (stroke.points.length < 2) return;

    ctx.strokeStyle = stroke.color;
    ctx.lineWidth = 4;
    ctx.lineCap = 'round';
    ctx.lineJoin = 'round';

    ctx.beginPath();
    ctx.moveTo(stroke.points[0].x, stroke.points[0].y);

    for (let i = 1; i < stroke.points.length; i++) {
      ctx.lineTo(stroke.points[i].x, stroke.points[i].y);
    }

    ctx.stroke();
  };

  const getCanvasPoint = (e: React.MouseEvent<HTMLCanvasElement>): Point => {
    const canvas = canvasRef.current;
    if (!canvas) return { x: 0, y: 0 };

    const rect = canvas.getBoundingClientRect();
    return {
      x: e.clientX - rect.left,
      y: e.clientY - rect.top,
    };
  };

  const handleMouseDown = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!allowDrawing) return;

    const point = getCanvasPoint(e);
    setIsDrawing(true);
    setCurrentStroke([point]);
    currentStrokeId.current = generateStrokeId();

    // Send drawing event using new format
    wsService.sendDrawingEvent(uuid, username, 'start', {
      strokeId: currentStrokeId.current,
      points: [point],
      color: selectedColor,
      lineWidth: 4
    });

    // Legacy format for backward compatibility
    wsService.sendMessage(uuid, 'draw', {
      type: 'start',
      color: selectedColor,
      point: { x: point.x, y: point.y },
    });
  };

  const handleMouseMove = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!allowDrawing || !isDrawing) return;

    const point = getCanvasPoint(e);
    setCurrentStroke((prev) => [...prev, point]);

    // Send drawing event using new format
    wsService.sendDrawingEvent(uuid, username, 'draw', {
      strokeId: currentStrokeId.current,
      points: [point],
      color: selectedColor,
      lineWidth: 4
    });

    // Legacy format for backward compatibility
    wsService.sendMessage(uuid, 'draw', {
      type: 'draw',
      point: { x: point.x, y: point.y },
    });
  };

  const handleMouseUp = () => {
    if (!allowDrawing || !isDrawing) return;

    setIsDrawing(false);
    
    // Send drawing event using new format
    wsService.sendDrawingEvent(uuid, username, 'end', {
      strokeId: currentStrokeId.current
    });

    // Legacy format for backward compatibility
    wsService.sendMessage(uuid, 'draw', {
      type: 'end',
    });

    // Add current stroke to completed strokes
    if (currentStroke.length > 0) {
      const newStroke: Stroke = {
        points: [...currentStroke],
        color: selectedColor,
      };
      setStrokes((prev) => {
        const updated = [...prev, newStroke];
        if (onStrokesChange) onStrokesChange(updated);
        return updated;
      });
      setCurrentStroke([]);
    }
    currentStrokeId.current = '';
  };

  const handleMouseLeave = () => {
    if (isDrawing) {
      handleMouseUp();
    }
  };

  return (
    <div style={{ display: 'inline-block', border: '2px solid #333' }}>
      <canvas
        ref={canvasRef}
        width={canvasWidth}
        height={canvasHeight}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseLeave}
        style={{
          cursor: allowDrawing ? 'crosshair' : 'not-allowed',
          display: 'block',
        }}
      />
    </div>
  );
};

export default DrawingCanvas;
