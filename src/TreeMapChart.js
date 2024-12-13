import React, { useState, useEffect } from 'react';
import { TreeMap, TreeMapSeries } from 'reaviz';

const TreeMapChart = () => {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedOption, setSelectedOption] = useState('account_ccy'); // Default is Currency

  // Handle change in dropdown selection
  const handleDropdownChange = (event) => {
    setSelectedOption(event.target.value); // Update state based on dropdown value
  };

  const fetchData = async (column) => {
    if (!column) {
      setError('Please select either Country or Currency');
      setLoading(false);
      return;
    }

    try {
      const response = await fetch('http://localhost:8080/heatmap/filterdata', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ filter: column }), // Send the selected column value as 'filter'
      });

      if (!response.ok) {
        throw new Error('Network response was not ok');
      }

      const result = await response.json();
      setData(result); // Set the response data for TreeMap
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Fetch data initially with 'account_ccy' (Currency) selected by default
  useEffect(() => {
    setLoading(true); // Show loading spinner
    fetchData('account_ccy'); // Initial API call when the component loads
  }, []); // This will only run once when the component mounts

  // Fetch data when the selected option changes
  useEffect(() => {
    setLoading(true); // Show loading spinner
    fetchData(selectedOption); // Fetch data for the selected filter
  }, [selectedOption]); // Re-run fetchData when selectedOption changes

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;

  return (
    <div>
      <div>
        <label htmlFor="dropdown">Group By: </label>
        <select id="dropdown" value={selectedOption} onChange={handleDropdownChange}>
          <option value="account_country">Country</option>
          <option value="account_ccy">Currency</option>
        </select>
      </div>

      <TreeMap
        height={600}
        width={1200}
        data={data}
        series={
          <TreeMapSeries
            colorScheme={[
              '#991f29',
              '#f23645',
              '#f77c80',
              '#c1c4cd',
              '#42bd7f',
              '#089950',
              '#056636',
            ]}
          />
        }
      />
    </div>
  );
};

export default TreeMapChart;
