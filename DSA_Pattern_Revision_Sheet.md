# DSA Pattern Revision Sheet (Pen & Paper Grind)

**Goal:** Internalize the *pattern*, not the question. For each pattern below: read it once, write the approach in your own words on paper, then code it cold in Notepad (no AI, no IDE autocomplete). Only after that, check it against an editor.

**How to use this file:**
- [ ] Topic understood conceptually
- [ ] Pattern explained out loud / on paper (your own words)
- [ ] Solved in Notepad without help
- [ ] Solved again 3 days later from memory (this is the actual test)

Target level: Product-based + FAANG+. Foundational topics kept lean; harder topics expanded into real interview-distinguishing patterns.

---

## 1. Arrays & Hashing (Foundational — keep lean)

| Pattern | Core Idea | Question |
|---|---|---|
| Frequency counting / hashmap lookup | Use a map to track counts or seen values in O(1) | [Two Sum](https://leetcode.com/problems/two-sum/) |
| Prefix sum | Precompute running sums to answer range/subarray queries fast | [Subarray Sum Equals K](https://leetcode.com/problems/subarray-sum-equals-k/) |
| In-place array manipulation | Swap/overwrite without extra space, often using a pointer to track write position | [Move Zeroes](https://leetcode.com/problems/move-zeroes/) |

---

## 2. Two Pointers

| Pattern | Core Idea | Question |
|---|---|---|
| Opposite-end pointers (sorted array) | Start at both ends, move based on comparison to target | [Two Sum II](https://leetcode.com/problems/two-sum-ii-input-array-is-sorted/) |
| Same-direction (fast/slow on array) | Both pointers move forward, one conditionally | [Remove Duplicates from Sorted Array](https://leetcode.com/problems/remove-duplicates-from-sorted-array/) |
| Container/area maximization | Shrink from the side with the limiting factor | [Container With Most Water](https://leetcode.com/problems/container-with-most-water/) |
| Three pointers (fix one, two-pointer the rest) | Fix an index, two-pointer the remaining subarray, skip duplicates | [3Sum](https://leetcode.com/problems/3sum/) |

---

## 3. Sliding Window

| Pattern | Core Idea | Question |
|---|---|---|
| Fixed-size window | Window size constant, slide and update sum/count incrementally | [Maximum Average Subarray I](https://leetcode.com/problems/maximum-average-subarray-i/) |
| Variable-size window (expand/shrink on condition) | Grow right pointer, shrink left pointer when condition breaks | [Longest Substring Without Repeating Characters](https://leetcode.com/problems/longest-substring-without-repeating-characters/) |
| Window with frequency map (anagram/substring match) | Track character counts in window, compare to target map | [Minimum Window Substring](https://leetcode.com/problems/minimum-window-substring/) |
| Monotonic deque in window | Maintain deque of indices to get max/min in window in O(1) | [Sliding Window Maximum](https://leetcode.com/problems/sliding-window-maximum/) |

---

## 4. Binary Search

| Pattern | Core Idea | Question |
|---|---|---|
| Standard search on sorted array | Classic lo/hi/mid template | [Binary Search](https://leetcode.com/problems/binary-search/) |
| Search in rotated sorted array | Identify which half is sorted, decide which side to search | [Search in Rotated Sorted Array](https://leetcode.com/problems/search-in-rotated-sorted-array/) |
| Binary search on answer space (not the array itself) | Binary search over possible answers, use a feasibility check function | [Koko Eating Bananas](https://leetcode.com/problems/koko-eating-bananas/) |
| Find boundary (first/last occurrence) | Modify standard BS to keep searching after match to find leftmost/rightmost | [Find First and Last Position of Element in Sorted Array](https://leetcode.com/problems/find-first-and-last-position-of-element-in-sorted-array/) |

---

## 5. Linked List

| Pattern | Core Idea | Question |
|---|---|---|
| Reversal (iterative, prev/curr/next) | Standard pointer-rewiring reversal | [Reverse Linked List](https://leetcode.com/problems/reverse-linked-list/) |
| Fast & slow pointers (cycle/middle detection) | Two pointers at different speeds | [Linked List Cycle](https://leetcode.com/problems/linked-list-cycle/) |
| Merge / merge-sort style on lists | Combine two sorted lists or split-merge recursively | [Merge Two Sorted Lists](https://leetcode.com/problems/merge-two-sorted-lists/) |
| Dummy node + multi-pointer manipulation | Use dummy head to simplify edge cases while removing/reordering nodes | [Remove Nth Node From End of List](https://leetcode.com/problems/remove-nth-node-from-end-of-list/) |

---

## 6. Stack & Queue

| Pattern | Core Idea | Question |
|---|---|---|
| Matching/balancing with stack | Push opening symbols, pop and validate on closing | [Valid Parentheses](https://leetcode.com/problems/valid-parentheses/) |
| Monotonic stack (next greater/smaller element) | Maintain increasing/decreasing stack, pop when violated | [Daily Temperatures](https://leetcode.com/problems/daily-temperatures/) |
| Min/Max stack design | Auxiliary stack tracks min/max alongside main stack | [Min Stack](https://leetcode.com/problems/min-stack/) |
| Queue via two stacks / stack via two queues | Simulate one structure using the other | [Implement Queue using Stacks](https://leetcode.com/problems/implement-queue-using-stacks/) |

---

## 7. Trees (Binary Tree / BST)

| Pattern | Core Idea | Question |
|---|---|---|
| DFS recursive traversal (pre/in/post) | Recursive function processing node then children | [Binary Tree Inorder Traversal](https://leetcode.com/problems/binary-tree-inorder-traversal/) |
| BFS level-order traversal | Queue-based, process level by level | [Binary Tree Level Order Traversal](https://leetcode.com/problems/binary-tree-level-order-traversal/) |
| Path sum / root-to-leaf accumulation | DFS carrying a running value, branch at each node | [Path Sum II](https://leetcode.com/problems/path-sum-ii/) |
| Lowest Common Ancestor (recursive bottom-up) | Return node if found in subtree, combine left/right results | [Lowest Common Ancestor of a Binary Tree](https://leetcode.com/problems/lowest-common-ancestor-of-a-binary-tree/) |
| BST property exploitation (search/insert/validate) | Use left < node < right to prune search space | [Validate Binary Search Tree](https://leetcode.com/problems/validate-binary-search-tree/) |
| Diameter / height-based DP on tree | Compute height recursively, update global answer using left+right heights | [Diameter of Binary Tree](https://leetcode.com/problems/diameter-of-binary-tree/) |
| Serialize/Deserialize (tree ↔ string) | Pre-order encode with null markers, recursive decode | [Serialize and Deserialize Binary Tree](https://leetcode.com/problems/serialize-and-deserialize-binary-tree/) |

---

## 8. Heaps / Priority Queue

| Pattern | Core Idea | Question |
|---|---|---|
| Top-K elements | Maintain heap of size K, push/pop to keep only the K best | [Kth Largest Element in an Array](https://leetcode.com/problems/kth-largest-element-in-an-array/) |
| Two heaps (median maintenance) | Max-heap for lower half, min-heap for upper half, balance sizes | [Find Median from Data Stream](https://leetcode.com/problems/find-median-from-data-stream/) |
| Merge K sorted structures | Heap of (value, source) pairs, pop min and push next from same source | [Merge k Sorted Lists](https://leetcode.com/problems/merge-k-sorted-lists/) |
| Greedy scheduling with heap | Heap tracks earliest-available/least-loaded resource | [Task Scheduler](https://leetcode.com/problems/task-scheduler/) |

---

## 9. Backtracking

| Pattern | Core Idea | Question |
|---|---|---|
| Subsets / combinations (include-exclude) | At each element, branch into "take it" / "don't take it" | [Subsets](https://leetcode.com/problems/subsets/) |
| Permutations (used[] tracking) | Track used elements, backtrack after each placement | [Permutations](https://leetcode.com/problems/permutations/) |
| Constraint satisfaction on grid (N-Queens style) | Place, validate constraints, recurse, undo | [N-Queens](https://leetcode.com/problems/n-queens/) |
| Combination sum (reuse vs no-reuse elements) | Recurse with/without advancing index depending on if reuse allowed | [Combination Sum](https://leetcode.com/problems/combination-sum/) |
| Word search / path building on grid | DFS with visited marking and backtrack (unmark) after exploring | [Word Search](https://leetcode.com/problems/word-search/) |

---

## 10. Graphs

| Pattern | Core Idea | Question |
|---|---|---|
| DFS/BFS connected components | Traverse and mark visited, count separate traversal starts | [Number of Islands](https://leetcode.com/problems/number-of-islands/) |
| Topological sort (Kahn's / DFS-based) | Track in-degrees, process zero in-degree nodes first | [Course Schedule](https://leetcode.com/problems/course-schedule/) |
| Union-Find (Disjoint Set) | Union connected nodes, find with path compression | [Number of Provinces](https://leetcode.com/problems/number-of-provinces/) |
| Shortest path unweighted (BFS) | BFS layer by layer, track distance/steps | [Word Ladder](https://leetcode.com/problems/word-ladder/) |
| Shortest path weighted (Dijkstra) | Min-heap pops closest unvisited node, relax neighbors | [Network Delay Time](https://leetcode.com/problems/network-delay-time/) |
| Cycle detection (directed vs undirected) | Directed: track recursion stack. Undirected: track parent | [Course Schedule II](https://leetcode.com/problems/course-schedule-ii/) |
| Multi-source BFS | Start BFS from all sources simultaneously, expand together | [Rotting Oranges](https://leetcode.com/problems/rotting-oranges/) |

---

## 11. Dynamic Programming

| Pattern | Core Idea | Question |
|---|---|---|
| 1D DP (Fibonacci-style, state = index) | dp[i] depends on dp[i-1], dp[i-2]... | [Climbing Stairs](https://leetcode.com/problems/climbing-stairs/) |
| 0/1 Knapsack | dp[i][w] = take or skip item i, track capacity used | [Partition Equal Subset Sum](https://leetcode.com/problems/partition-equal-subset-sum/) |
| Unbounded Knapsack (coin/item reuse allowed) | Same as knapsack but don't decrement item index on take | [Coin Change](https://leetcode.com/problems/coin-change/) |
| Longest Common Subsequence (2D string DP) | dp[i][j] from two strings, match or skip | [Longest Common Subsequence](https://leetcode.com/problems/longest-common-subsequence/) |
| Longest Increasing Subsequence | dp[i] = best LIS ending at i, or binary search optimization | [Longest Increasing Subsequence](https://leetcode.com/problems/longest-increasing-subsequence/) |
| Grid path DP | dp[r][c] built from top/left neighbors | [Unique Paths](https://leetcode.com/problems/unique-paths/) |
| Interval DP | dp[i][j] over a range, decided by a split point k inside | [Burst Balloons](https://leetcode.com/problems/burst-balloons/) |
| DP on trees | Combine dp results of children at each node | [House Robber III](https://leetcode.com/problems/house-robber-iii/) |
| State machine DP (buy/sell/hold style) | Track multiple states per index (holding vs not holding) | [Best Time to Buy and Sell Stock with Cooldown](https://leetcode.com/problems/best-time-to-buy-and-sell-stock-with-cooldown/) |

---

## 12. Greedy

| Pattern | Core Idea | Question |
|---|---|---|
| Interval scheduling (sort by end time) | Sort intervals, greedily pick non-overlapping ones | [Non-overlapping Intervals](https://leetcode.com/problems/non-overlapping-intervals/) |
| Greedy + sorting for resource allocation | Sort both arrays, match greedily | [Assign Cookies](https://leetcode.com/problems/assign-cookies/) |
| Jump game / reachability greedy | Track farthest reachable index while iterating | [Jump Game](https://leetcode.com/problems/jump-game/) |

---

## 13. Tries

| Pattern | Core Idea | Question |
|---|---|---|
| Prefix tree insert/search | Node has children map + end-of-word flag | [Implement Trie (Prefix Tree)](https://leetcode.com/problems/implement-trie-prefix-tree/) |
| Word search with trie + backtracking (combo pattern) | Build trie of words, DFS grid while walking trie simultaneously | [Word Search II](https://leetcode.com/problems/word-search-ii/) |

---

## 14. Bit Manipulation (Foundational — keep lean)

| Pattern | Core Idea | Question |
|---|---|---|
| XOR trick for single/unique element | XOR cancels duplicates, leaves the unique one | [Single Number](https://leetcode.com/problems/single-number/) |
| Bitmasking for subsets/states | Represent subset/state as integer bitmask | [Subsets](https://leetcode.com/problems/subsets/) (bitmask variant) |

---

## 15. Common "Combo" Patterns Interviewers Love (Top-tier specific)

These blend two topics — common at FAANG+ to see if you can compose patterns, not just recall them.

| Pattern | Core Idea | Question |
|---|---|---|
| DFS + Memoization (top-down DP on grid/graph) | Recursive DFS with a memo dict to avoid recomputation | [Longest Increasing Path in a Matrix](https://leetcode.com/problems/longest-increasing-path-in-a-matrix/) |
| Sliding window + hashmap (substring constraints) | Window tracks counts, shrink when constraint violated | [Longest Repeating Character Replacement](https://leetcode.com/problems/longest-repeating-character-replacement/) |
| Binary search + greedy check function | BS over answer, greedy/simulation function validates feasibility | [Split Array Largest Sum](https://leetcode.com/problems/split-array-largest-sum/) |
| Union-Find + sorting (Kruskal's-style) | Sort edges by weight, union if not already connected | [Min Cost to Connect All Points](https://leetcode.com/problems/min-cost-to-connect-all-points/) |
| Heap + greedy (interval/meeting room style) | Heap tracks earliest end time among active intervals | [Meeting Rooms II](https://leetcode.com/problems/meeting-rooms-ii/) |

---

## How to actually use this for your plan

1. **Pick one topic per day or two** depending on how foundational vs new it is to you.
2. **Pen & paper first:** for each pattern, write 3-4 lines — what's the trigger to recognize this pattern, what's the core data structure/technique, what's the time complexity.
3. **Notepad second:** code the representative question with zero AI help. Time yourself (aim for 20-25 min for mediums).
4. **Revisit after 3 days:** re-solve the same question from memory. If you can't, the pattern hasn't stuck yet — don't move on.
5. **Mock interview simulation (weeks 3-4):** once patterns are solid, do timed mock problems out loud, explaining approach before coding — this rebuilds the "talk while typing" muscle that AI assistance let you skip.

Total: **~50 core patterns** across 15 topics. This is deliberately a finite, closed set — once you've internalized these, almost any interview question is a recombination or minor variant of one of them.
