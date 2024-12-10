import React, { useState } from 'react';
import TreeMapChart from './TreeMapChart';

function App() {
  const [groupBy, setGroupBy] = useState('COUNTRY');


  return (
    <div style={{ 
      padding: '20px',
      maxWidth: '1200px',
      margin: '0 auto',
      fontFamily: 'Arial, sans-serif'
    }}>
      <TreeMapChart groupBy={groupBy} />
    </div>
  );
}

export default App;