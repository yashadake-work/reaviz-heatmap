import React from 'react';
import { TreeMap, TreeMapSeries } from 'reaviz';

const TreeMapChart = () => (
  <TreeMap
    height={600}
    width={400}
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
    data={[
      {
        key: 'USA',
        data: [
          { key: 'AccNo1', data: 20 },
          { key: 'AccNo2', data: 30 },
          { key: 'AccNo3', data: 10 },
          { key: 'AccNo4', data: 20 },
          { key: 'AccNo5', data: 30 },
          { key: 'AccNo6', data: 10 }
        ]
      },
      {
        key: 'INDIA',
        data: [
          { key: 'AccNo1', data: 15 },
          { key: 'AccNo2', data: 25 },
          { key: 'AccNo3', data: 15 },
          { key: 'AccNo4', data: 30 },
          { key: 'AccNo5', data: 25 },
          { key: 'AccNo6', data: 15 }
        ]
      },
      {
        key: 'RUSSIA',
        data: [
          { key: 'AccNo1', data: 40 },
          { key: 'AccNo2', data: 15 },
          { key: 'AccNo3', data: 20 },
          { key: 'AccNo4', data: 35 },
          { key: 'AccNo5', data: 15 },
          { key: 'AccNo6', data: 40 },
          { key: 'AccNo7', data: 15 },
          { key: 'AccNo8', data: 13 },
          { key: 'AccNo9', data: 19 }
        ]
      },
      {
        key: 'SPAIN',
        data: [
          { key: 'AccNo1', data: 40 },
          { key: 'AccNo2', data: 15 },
          { key: 'AccNo3', data: 25 },
          { key: 'AccNo4', data: 5 }
        ]
      },
      {
        key: 'GERMANY',
        data: [
          { key: 'AccNo1', data: 40 },
          { key: 'AccNo2', data: 15 },
          { key: 'AccNo3', data: 15 }
        ]
      },
      {
        key: 'FRANCE',
        data: [
          { key: 'AccNo1', data: 50 },
          { key: 'AccNo2', data: 30 }
        ]
      },
      {
        key: 'POLAND',
        data: [
          { key: 'AccNo1', data: 15 },
          { key: 'AccNo2', data: 25 },
          { key: 'AccNo3', data: 15 }
        ]
      },
      {
        key: 'MALASHIA',
        data: [
          { key: 'AccNo1', data: 60 }
        ]
      },
      {
        key: 'ENGLAND',
        data: [
          { key: 'AccNo1', data: [{ key: 'AccNo8 ', data: 15 },
            { key: 'AccNo8', data: 15 }] },
          { key: 'AccNo2', data: 15 },
          { key: 'AccNo3', data: 15 }
        ]
      }
    ]}
  />
);

export default TreeMapChart;