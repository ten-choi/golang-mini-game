import React from 'react';

interface LoadingSpinnerProps {
  size?: 'small' | 'medium' | 'large';
  message?: string;
}

/**
 * Reusable Loading Spinner Component
 */
export const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({ 
  size = 'medium', 
  message 
}) => {
  const sizeMap = {
    small: 30,
    medium: 50,
    large: 70,
  };

  const spinnerSize = sizeMap[size];

  return (
    <div style={styles.container}>
      <div 
        style={{
          ...styles.spinner,
          width: spinnerSize,
          height: spinnerSize,
        }}
      />
      {message && <p style={styles.message}>{message}</p>}
    </div>
  );
};

const styles = {
  container: {
    display: 'flex',
    flexDirection: 'column' as const,
    alignItems: 'center',
    justifyContent: 'center',
    padding: '20px',
  },
  spinner: {
    border: '4px solid rgba(0, 0, 0, 0.1)',
    borderTop: '4px solid #3498db',
    borderRadius: '50%',
    animation: 'spin 1s linear infinite',
  },
  message: {
    marginTop: '16px',
    fontSize: '14px',
    color: '#666',
  },
};

// Add keyframe animation to global styles
if (typeof document !== 'undefined') {
  const styleSheet = document.styleSheets[0];
  const keyframes = `
    @keyframes spin {
      0% { transform: rotate(0deg); }
      100% { transform: rotate(360deg); }
    }
  `;
  
  try {
    styleSheet.insertRule(keyframes, styleSheet.cssRules.length);
  } catch (e) {
    // Animation might already exist
  }
}
