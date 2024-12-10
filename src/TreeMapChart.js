import React, { useState, useEffect } from 'react';
import { TreeMap, TreeMapSeries } from 'reaviz';

const TreeMapChart = () => {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/treedata');
        if (!response.ok) {
          throw new Error('Network response was not ok');
        }
        const result = await response.json();
        setData(result);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;

  return (
    <TreeMap
      height={600}
      width={1200}
      series={
        <TreeMapSeries
          colorScheme={[
            '#991f29',
            '#f23645',
            '#f77c80',
            '#c1c4cd',
            '#42bd7f',
            '#089950',
            '#056636'
          ]}
        />
      }
      data={data}
    />
  );
};

export default TreeMapChart;