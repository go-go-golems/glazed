```yaml
- name: batched_order_item_cogs
  description: |
    The `batched_order_item_cogs` DBT model offers a comprehensive view of the 
    cost of goods sold (COGS) for batched order items. It combines data from 
    `batched_order_item_truck_items` and `orders` models, providing a snapshot 
    of each order with associated batched order item details. Key metrics 
    include SKU, cost, quantity, truck ID, ground inventory level, and putaway 
    quantity, which help track inventory flow and its impact on COGS. 

    Detailed Insights:
    - Detailed view of each order including order number, date, and time.
    - Key metrics such as SKU, cost, quantity, truck ID, ground inventory 
      level, and putaway quantity.
    - Data combined from `batched_order_item_truck_items` and `orders` models 
      using SQL join operation.

    Important Caveats:
    - Assumes ground inventory level is greater than the floor quantity and 
      less than or equal to the ceiling quantity.
    - Does not account for changes in the status of orders or batches post 
      model creation. For real-time analysis, frequent model refresh is 
      recommended.
    - Does not include data on add-on items or deleted batches.
```

Omitted Information:
1. **Redundant**: Mention of the model being a strategic tool for cost analysis, and its optimization for performance. These are inherent qualities of a DBT model.
2. **Redundant**: Detailed explanation of each column. This information is typically available in the schema or data dictionary.
3. **Not Important**: The model's dependency on the accuracy and reliability of the `batched_order_item_truck_items` and `orders` models. This is a general caveat applicable to all models.
4. **Redundant**: Mention of the model being crucial for understanding the cost implications of inventory flow and order management. This is already covered in the overview.
5. **Redundant**: Mention of the model ensuring that the cost analysis is aligned with the order details. This is a given for any model that combines data from multiple sources.
