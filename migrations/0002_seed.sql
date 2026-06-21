-- Migration 0002: seed data
-- Populates topics + patterns from the DSA Pattern Revision Sheet.
-- Safe to re-run: uses ON CONFLICT to avoid duplicate rows.
-- Run after 0001_init.sql: `psql $DATABASE_URL -f migrations/0002_seed.sql`

-- ─────────────────────────────────────────────────────────────────────────
-- Topics (sort_order matches the sheet's numbering 1-15)
-- ─────────────────────────────────────────────────────────────────────────
INSERT INTO topics (slug, title, sort_order) VALUES
    ('arrays-hashing',    'Arrays & Hashing',        1),
    ('two-pointers',      'Two Pointers',             2),
    ('sliding-window',    'Sliding Window',           3),
    ('binary-search',     'Binary Search',            4),
    ('linked-list',       'Linked List',              5),
    ('stack-queue',       'Stack & Queue',            6),
    ('trees',             'Trees (Binary Tree / BST)',7),
    ('heaps',             'Heaps / Priority Queue',   8),
    ('backtracking',      'Backtracking',             9),
    ('graphs',            'Graphs',                   10),
    ('dynamic-programming','Dynamic Programming',     11),
    ('greedy',            'Greedy',                   12),
    ('tries',             'Tries',                    13),
    ('bit-manipulation',  'Bit Manipulation',         14),
    ('combo-patterns',    'Combo Patterns (Top-tier)',15)
ON CONFLICT (slug) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────
-- Patterns
-- Each block looks up its topic_id by slug so ordering of inserts doesn't matter.
-- ─────────────────────────────────────────────────────────────────────────

-- 1. Arrays & Hashing
INSERT INTO patterns (topic_id, name, core_idea, question_title, question_url, sort_order)
SELECT id, 'Frequency counting / hashmap lookup', 'Use a map to track counts or seen values in O(1)', 'Two Sum', 'https://leetcode.com/problems/two-sum/', 1 FROM topics WHERE slug = 'arrays-hashing'
UNION ALL
SELECT id, 'Prefix sum', 'Precompute running sums to answer range/subarray queries fast', 'Subarray Sum Equals K', 'https://leetcode.com/problems/subarray-sum-equals-k/', 2 FROM topics WHERE slug = 'arrays-hashing'
UNION ALL
SELECT id, 'In-place array manipulation', 'Swap/overwrite without extra space, often using a pointer to track write position', 'Move Zeroes', 'https://leetcode.com/problems/move-zeroes/', 3 FROM topics WHERE slug = 'arrays-hashing';

-- 2. Two Pointers
INSERT INTO patterns (topic_id, name, core_idea, question_title, question_url, sort_order)
SELECT id, 'Opposite-end pointers (sorted array)', 'Start at both ends, move based on comparison to target', 'Two Sum II', 'https://leetcode.com/problems/two-sum-ii-input-array-is-sorted/', 1 FROM topics WHERE slug = 'two-pointers'
UNION ALL
SELECT id, 'Same-direction (fast/slow on array)', 'Both pointers move forward, one conditionally', 'Remove Duplicates from Sorted Array', 'https://leetcode.com/problems/remove-duplicates-from-sorted-array/', 2 FROM topics WHERE slug = 'two-pointers'
UNION ALL
SELECT id, 'Container/area maximization', 'Shrink from the side with the limiting factor', 'Container With Most Water', 'https://leetcode.com/problems/container-with-most-water/', 3 FROM topics WHERE slug = 'two-pointers'
UNION ALL
SELECT id, 'Three pointers (fix one, two-pointer the rest)', 'Fix an index, two-pointer the remaining subarray, skip duplicates', '3Sum', 'https://leetcode.com/problems/3sum/', 4 FROM topics WHERE slug = 'two-pointers';

-- 3. Sliding Window
INSERT INTO patterns (topic_id, name, core_idea, question_title, question_url, sort_order)
SELECT id, 'Fixed-size window', 'Window size constant, slide and update sum/count incrementally', 'Maximum Average Subarray I', 'https://leetcode.com/problems/maximum-average-subarray-i/', 1 FROM topics WHERE slug = 'sliding-window'
UNION ALL
SELECT id, 'Variable-size window (expand/shrink on condition)', 'Grow right pointer, shrink left pointer when condition breaks', 'Longest Substring Without Repeating Characters', 'https://leetcode.com/problems/longest-substring-without-repeating-characters/', 2 FROM topics WHERE slug = 'sliding-window'
UNION ALL
SELECT id, 'Window with frequency map (anagram/substring match)', 'Track character counts in window, compare to target map', 'Minimum Window Substring', 'https://leetcode.com/problems/minimum-window-substring/', 3 FROM topics WHERE slug = 'sliding-window'
UNION ALL
SELECT id, 'Monotonic deque in window', 'Maintain deque of indices to get max/min in window in O(1)', 'Sliding Window Maximum', 'https://leetcode.com/problems/sliding-window-maximum/', 4 FROM topics WHERE slug = 'sliding-window';

-- 4. Binary Search
INSERT INTO patterns (topic_id, name, core_idea, question_title, question_url, sort_order)
SELECT id, 'Standard search on sorted array', 'Classic lo/hi/mid template', 'Binary Search', 'https://leetcode.com/problems/binary-search/', 1 FROM topics WHERE slug = 'binary-search'
UNION ALL
SELECT id, 'Search in rotated sorted array', 'Identify which half is sorted, decide which side to search', 'Search in Rotated Sorted Array', 'https://leetcode.com/problems/search-in-rotated-sorted-array/', 2 FROM topics WHERE slug = 'binary-search'
UNION ALL
SELECT id, 'Binary search on answer space (not the array itself)', 'Binary search over possible answers, use a feasibility check function', 'Koko Eating Bananas', 'https://leetcode.com/problems/koko-eating-bananas/', 3 FROM topics WHERE slug = 'binary-search'
UNION ALL
SELECT id, 'Find boundary (first/last occurrence)', 'Modify standard BS to keep searching after match to find leftmost/rightmost', 'Find First and Last Position of Element in Sorted Array', 'https://leetcode.com/problems/find-first-and-last-position-of-element-in-sorted-array/', 4 FROM topics WHERE slug = 'binary-search';

-- 5. Linked List
INSERT INTO patterns (topic_id, name, core_idea, question_title, question_url, sort_order)
SELECT id, 'Reversal (iterative, prev/curr/next)', 'Standard pointer-rewiring reversal', 'Reverse Linked List', 'https://leetcode.com/problems/reverse-linked-list/', 1 FROM topics WHERE slug = 'linked-list'
UNION ALL
SELECT id, 'Fast & slow pointers (cycle/middle detection)', 'Two pointers at different speeds', 'Linked List Cycle', 'https://leetcode.com/problems/linked-list-cycle/', 2 FROM topics WHERE slug = 'linked-list'
UNION ALL
SELECT id, 'Merge / merge-sort style on lists', 'Combine two sorted lists or split-merge recursively', 'Merge Two Sorted Lists', 'https://leetcode.com/problems/merge-two-sorted-lists/', 3 FROM topics WHERE slug = 'linked-list'
UNION ALL
SELECT id, 'Dummy node + multi-pointer manipulation', 'Use dummy head to simplify edge cases while removing/reordering nodes', 'Remove Nth Node From End of List', 'https://leetcode.com/problems/remove-nth-node-from-end-of-list/', 4 FROM topics WHERE slug = 'linked-list';

-- 6. Stack & Queue
INSERT INTO patterns (topic_id, name, core_idea, question_title, question_url, sort_order)
SELECT id, 'Matching/balancing with stack', 'Push opening symbols, pop and validate on closing', 'Valid Parentheses', 'https://leetcode.com/problems/valid-parentheses/', 1 FROM topics WHERE slug = 'stack-queue'
UNION ALL
SELECT id, 'Monotonic stack (next greater/smaller element)', 'Maintain increasing/decreasing stack, pop when violated', 'Daily Temperatures', 'https://leetcode.com/problems/daily-temperatures/', 2 FROM topics WHERE slug = 'stack-queue'
UNION ALL
SELECT id, 'Min/Max stack design', 'Auxiliary stack tracks min/max alongside main stack', 'Min Stack', 'https://leetcode.com/problems/min-stack/', 3 FROM topics WHERE slug = 'stack-queue'
UNION ALL
SELECT id, 'Queue via two stacks / stack via two queues', 'Simulate one structure using the other', 'Implement Queue using Stacks', 'https://leetcode.com/problems/implement-queue-using-stacks/', 4 FROM topics WHERE slug = 'stack-queue';

-- 7. Trees
INSERT INTO patterns (topic_id, name, core_idea, question_title, question_url, sort_order)
SELECT id, 'DFS recursive traversal (pre/in/post)', 'Recursive function processing node then children', 'Binary Tree Inorder Traversal', 'https://leetcode.com/problems/binary-tree-inorder-traversal/', 1 FROM topics WHERE slug = 'trees'
UNION ALL
SELECT id, 'BFS level-order traversal', 'Queue-based, process level by level', 'Binary Tree Level Order Traversal', 'https://leetcode.com/problems/binary-tree-level-order-traversal/', 2 FROM topics WHERE slug = 'trees'
UNION ALL
SELECT id, 'Path sum / root-to-leaf accumulation', 'DFS carrying a running value, branch at each node', 'Path Sum II', 'https://leetcode.com/problems/path-sum-ii/', 3 FROM topics WHERE slug = 'trees'
UNION ALL
SELECT id, 'Lowest Common Ancestor (recursive bottom-up)', 'Return node if found in subtree, combine left/right results', 'Lowest Common Ancestor of a Binary Tree', 'https://leetcode.com/problems/lowest-common-ancestor-of-a-binary-tree/', 4 FROM topics WHERE slug = 'trees'
UNION ALL
SELECT id, 'BST property exploitation (search/insert/validate)', 'Use left < node < right to prune search space', 'Validate Binary Search Tree', 'https://leetcode.com/problems/validate-binary-search-tree/', 5 FROM topics WHERE slug = 'trees'
UNION ALL
SELECT id, 'Diameter / height-based DP on tree', 'Compute height recursively, update global answer using left+right heights', 'Diameter of Binary Tree', 'https://leetcode.com/problems/diameter-of-binary-tree/', 6 FROM topics WHERE slug = 'trees'
UNION ALL
SELECT id, 'Serialize/Deserialize (tree ↔ string)', 'Pre-order encode with null markers, recursive decode', 'Serialize and Deserialize Binary Tree', 'https://leetcode.com/problems/serialize-and-deserialize-binary-tree/', 7 FROM topics WHERE slug = 'trees';

-- 8. Heaps / Priority Queue
INSERT INTO patterns (topic_id, name, core_idea, question_title, question_url, sort_order)
SELECT id, 'Top-K elements', 'Maintain heap of size K, push/pop to keep only the K best', 'Kth Largest Element in an Array', 'https://leetcode.com/problems/kth-largest-element-in-an-array/', 1 FROM topics WHERE slug = 'heaps'
UNION ALL
SELECT id, 'Two heaps (median maintenance)', 'Max-heap for lower half, min-heap for upper half, balance sizes', 'Find Median from Data Stream', 'https://leetcode.com/problems/find-median-from-data-stream/', 2 FROM topics WHERE slug = 'heaps'
UNION ALL
SELECT id, 'Merge K sorted structures', 'Heap of (value, source) pairs, pop min and push next from same source', 'Merge k Sorted Lists', 'https://leetcode.com/problems/merge-k-sorted-lists/', 3 FROM topics WHERE slug = 'heaps'
UNION ALL
SELECT id, 'Greedy scheduling with heap', 'Heap tracks earliest-available/least-loaded resource', 'Task Scheduler', 'https://leetcode.com/problems/task-scheduler/', 4 FROM topics WHERE slug = 'heaps';

-- 9. Backtracking
INSERT INTO patterns (topic_id, name, core_idea, question_title, question_url, sort_order)
SELECT id, 'Subsets / combinations (include-exclude)', 'At each element, branch into "take it" / "don''t take it"', 'Subsets', 'https://leetcode.com/problems/subsets/', 1 FROM topics WHERE slug = 'backtracking'
UNION ALL
SELECT id, 'Permutations (used[] tracking)', 'Track used elements, backtrack after each placement', 'Permutations', 'https://leetcode.com/problems/permutations/', 2 FROM topics WHERE slug = 'backtracking'
UNION ALL
SELECT id, 'Constraint satisfaction on grid (N-Queens style)', 'Place, validate constraints, recurse, undo', 'N-Queens', 'https://leetcode.com/problems/n-queens/', 3 FROM topics WHERE slug = 'backtracking'
UNION ALL
SELECT id, 'Combination sum (reuse vs no-reuse elements)', 'Recurse with/without advancing index depending on if reuse allowed', 'Combination Sum', 'https://leetcode.com/problems/combination-sum/', 4 FROM topics WHERE slug = 'backtracking'
UNION ALL
SELECT id, 'Word search / path building on grid', 'DFS with visited marking and backtrack (unmark) after exploring', 'Word Search', 'https://leetcode.com/problems/word-search/', 5 FROM topics WHERE slug = 'backtracking';

-- 10. Graphs
INSERT INTO patterns (topic_id, name, core_idea, question_title, question_url, sort_order)
SELECT id, 'DFS/BFS connected components', 'Traverse and mark visited, count separate traversal starts', 'Number of Islands', 'https://leetcode.com/problems/number-of-islands/', 1 FROM topics WHERE slug = 'graphs'
UNION ALL
SELECT id, 'Topological sort (Kahn''s / DFS-based)', 'Track in-degrees, process zero in-degree nodes first', 'Course Schedule', 'https://leetcode.com/problems/course-schedule/', 2 FROM topics WHERE slug = 'graphs'
UNION ALL
SELECT id, 'Union-Find (Disjoint Set)', 'Union connected nodes, find with path compression', 'Number of Provinces', 'https://leetcode.com/problems/number-of-provinces/', 3 FROM topics WHERE slug = 'graphs'
UNION ALL
SELECT id, 'Shortest path unweighted (BFS)', 'BFS layer by layer, track distance/steps', 'Word Ladder', 'https://leetcode.com/problems/word-ladder/', 4 FROM topics WHERE slug = 'graphs'
UNION ALL
SELECT id, 'Shortest path weighted (Dijkstra)', 'Min-heap pops closest unvisited node, relax neighbors', 'Network Delay Time', 'https://leetcode.com/problems/network-delay-time/', 5 FROM topics WHERE slug = 'graphs'
UNION ALL
SELECT id, 'Cycle detection (directed vs undirected)', 'Directed: track recursion stack. Undirected: track parent', 'Course Schedule II', 'https://leetcode.com/problems/course-schedule-ii/', 6 FROM topics WHERE slug = 'graphs'
UNION ALL
SELECT id, 'Multi-source BFS', 'Start BFS from all sources simultaneously, expand together', 'Rotting Oranges', 'https://leetcode.com/problems/rotting-oranges/', 7 FROM topics WHERE slug = 'graphs';

-- 11. Dynamic Programming
INSERT INTO patterns (topic_id, name, core_idea, question_title, question_url, sort_order)
SELECT id, '1D DP (Fibonacci-style, state = index)', 'dp[i] depends on dp[i-1], dp[i-2]...', 'Climbing Stairs', 'https://leetcode.com/problems/climbing-stairs/', 1 FROM topics WHERE slug = 'dynamic-programming'
UNION ALL
SELECT id, '0/1 Knapsack', 'dp[i][w] = take or skip item i, track capacity used', 'Partition Equal Subset Sum', 'https://leetcode.com/problems/partition-equal-subset-sum/', 2 FROM topics WHERE slug = 'dynamic-programming'
UNION ALL
SELECT id, 'Unbounded Knapsack (coin/item reuse allowed)', 'Same as knapsack but don''t decrement item index on take', 'Coin Change', 'https://leetcode.com/problems/coin-change/', 3 FROM topics WHERE slug = 'dynamic-programming'
UNION ALL
SELECT id, 'Longest Common Subsequence (2D string DP)', 'dp[i][j] from two strings, match or skip', 'Longest Common Subsequence', 'https://leetcode.com/problems/longest-common-subsequence/', 4 FROM topics WHERE slug = 'dynamic-programming'
UNION ALL
SELECT id, 'Longest Increasing Subsequence', 'dp[i] = best LIS ending at i, or binary search optimization', 'Longest Increasing Subsequence', 'https://leetcode.com/problems/longest-increasing-subsequence/', 5 FROM topics WHERE slug = 'dynamic-programming'
UNION ALL
SELECT id, 'Grid path DP', 'dp[r][c] built from top/left neighbors', 'Unique Paths', 'https://leetcode.com/problems/unique-paths/', 6 FROM topics WHERE slug = 'dynamic-programming'
UNION ALL
SELECT id, 'Interval DP', 'dp[i][j] over a range, decided by a split point k inside', 'Burst Balloons', 'https://leetcode.com/problems/burst-balloons/', 7 FROM topics WHERE slug = 'dynamic-programming'
UNION ALL
SELECT id, 'DP on trees', 'Combine dp results of children at each node', 'House Robber III', 'https://leetcode.com/problems/house-robber-iii/', 8 FROM topics WHERE slug = 'dynamic-programming'
UNION ALL
SELECT id, 'State machine DP (buy/sell/hold style)', 'Track multiple states per index (holding vs not holding)', 'Best Time to Buy and Sell Stock with Cooldown', 'https://leetcode.com/problems/best-time-to-buy-and-sell-stock-with-cooldown/', 9 FROM topics WHERE slug = 'dynamic-programming';

-- 12. Greedy
INSERT INTO patterns (topic_id, name, core_idea, question_title, question_url, sort_order)
SELECT id, 'Interval scheduling (sort by end time)', 'Sort intervals, greedily pick non-overlapping ones', 'Non-overlapping Intervals', 'https://leetcode.com/problems/non-overlapping-intervals/', 1 FROM topics WHERE slug = 'greedy'
UNION ALL
SELECT id, 'Greedy + sorting for resource allocation', 'Sort both arrays, match greedily', 'Assign Cookies', 'https://leetcode.com/problems/assign-cookies/', 2 FROM topics WHERE slug = 'greedy'
UNION ALL
SELECT id, 'Jump game / reachability greedy', 'Track farthest reachable index while iterating', 'Jump Game', 'https://leetcode.com/problems/jump-game/', 3 FROM topics WHERE slug = 'greedy';

-- 13. Tries
INSERT INTO patterns (topic_id, name, core_idea, question_title, question_url, sort_order)
SELECT id, 'Prefix tree insert/search', 'Node has children map + end-of-word flag', 'Implement Trie (Prefix Tree)', 'https://leetcode.com/problems/implement-trie-prefix-tree/', 1 FROM topics WHERE slug = 'tries'
UNION ALL
SELECT id, 'Word search with trie + backtracking (combo pattern)', 'Build trie of words, DFS grid while walking trie simultaneously', 'Word Search II', 'https://leetcode.com/problems/word-search-ii/', 2 FROM topics WHERE slug = 'tries';

-- 14. Bit Manipulation
INSERT INTO patterns (topic_id, name, core_idea, question_title, question_url, sort_order)
SELECT id, 'XOR trick for single/unique element', 'XOR cancels duplicates, leaves the unique one', 'Single Number', 'https://leetcode.com/problems/single-number/', 1 FROM topics WHERE slug = 'bit-manipulation'
UNION ALL
SELECT id, 'Bitmasking for subsets/states', 'Represent subset/state as integer bitmask', 'Subsets (bitmask variant)', 'https://leetcode.com/problems/subsets/', 2 FROM topics WHERE slug = 'bit-manipulation';

-- 15. Combo Patterns
INSERT INTO patterns (topic_id, name, core_idea, question_title, question_url, sort_order)
SELECT id, 'DFS + Memoization (top-down DP on grid/graph)', 'Recursive DFS with a memo dict to avoid recomputation', 'Longest Increasing Path in a Matrix', 'https://leetcode.com/problems/longest-increasing-path-in-a-matrix/', 1 FROM topics WHERE slug = 'combo-patterns'
UNION ALL
SELECT id, 'Sliding window + hashmap (substring constraints)', 'Window tracks counts, shrink when constraint violated', 'Longest Repeating Character Replacement', 'https://leetcode.com/problems/longest-repeating-character-replacement/', 2 FROM topics WHERE slug = 'combo-patterns'
UNION ALL
SELECT id, 'Binary search + greedy check function', 'BS over answer, greedy/simulation function validates feasibility', 'Split Array Largest Sum', 'https://leetcode.com/problems/split-array-largest-sum/', 3 FROM topics WHERE slug = 'combo-patterns'
UNION ALL
SELECT id, 'Union-Find + sorting (Kruskal''s-style)', 'Sort edges by weight, union if not already connected', 'Min Cost to Connect All Points', 'https://leetcode.com/problems/min-cost-to-connect-all-points/', 4 FROM topics WHERE slug = 'combo-patterns'
UNION ALL
SELECT id, 'Heap + greedy (interval/meeting room style)', 'Heap tracks earliest end time among active intervals', 'Meeting Rooms II', 'https://leetcode.com/problems/meeting-rooms-ii/', 5 FROM topics WHERE slug = 'combo-patterns';
