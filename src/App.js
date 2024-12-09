import React, { useState } from 'react';
import TreeMapChart from './TreeMapChart';

function App() {
  const [groupBy, setGroupBy] = useState('COUNTRY');

  const handleChange = (event) => {
    setGroupBy(event.target.value);
  };

  return (
    <div style={{ 
      padding: '20px',
      maxWidth: '1200px',
      margin: '0 auto',
      fontFamily: 'Arial, sans-serif'
    }}>
      <label htmlFor="groupBy">Group By:</label>
      <select id="groupBy" value={groupBy} onChange={handleChange}>
        <option value="COUNTRY">COUNTRY</option>
        <option value="CURRENCY">CURRENCY</option>
      </select>
      <TreeMapChart groupBy={groupBy} />
    </div>
  );
}

export default App;