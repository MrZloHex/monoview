## Basic Reachability Analysis for Linear Systems

This note summarizes core formulas and tests for reachability (a.k.a. controllability over a finite horizon) of linear time-invariant (LTI) systems, in both continuous and discrete time. All formulas use block math with $$...$$.

### System models

- Continuous time LTI:
  $$
  \dot x(t) = A x(t) + B u(t), \qquad x(0) = x_0.
  $$
- Discrete time LTI:
  $$
  x_{k+1} = A x_k + B u_k, \qquad x_0\ \text{given}.
  $$

A state \(x_T\) (or \(x_N\)) is reachable if there exists an input that drives the system from the initial state to that target in finite time.

### Continuous time: reachable set, Gramian, and minimal energy

State at time \(T\):
$$
x(T) = e^{AT} x_0 + \int_{0}^{T} e^{A(T-\tau)}\, B\, u(\tau)\, d\tau.
$$

- With \(x_0 = 0\), the set of reachable states at time \(T\) is
  $$
  \mathcal{R}(T) = \left\{ \int_{0}^{T} e^{A(T-\tau)} B u(\tau)\, d\tau : u\ \text{square integrable} \right\}.
  $$
- Define the finite-horizon reachability Gramian
  $$
  W_r(T) = \int_{0}^{T} e^{A\tau} B B^{\top} e^{A^{\top} \tau} \, d\tau.
  $$
  Then \(x_T\) (from \(x_0=0\)) is reachable with finite energy iff \(x_T \in \operatorname{range}(W_r(T))\). If \(W_r(T)\) is positive definite, every state is reachable at time \(T\).

- Minimal-energy input to reach \(x_T\) from \(x_0=0\):
  $$
  u^{\star}(t) = B^{\top} e^{A^{\top} (T - t)}\, W_r(T)^{-1}\, x_T.
  $$
  The minimal control energy is
  $$
  J^{\star} = \int_{0}^{T} \lVert u^{\star}(t) \rVert^2 dt = x_T^{\top} W_r(T)^{-1} x_T.
  $$

- Infinite-horizon Gramian (exists when \(A\) is Hurwitz):
  $$
  A W_c + W_c A^{\top} + B B^{\top} = 0, \qquad W_c = \int_{0}^{\infty} e^{A\tau} B B^{\top} e^{A^{\top} \tau} \, d\tau.
  $$

### Discrete time: reachable set, matrix, Gramian, and minimal energy

State after \(N\) steps:
$$
x_N = A^{N} x_0 + \sum_{i=0}^{N-1} A^{N-1-i} B\, u_i.
$$

- With \(x_0 = 0\), the reachable set at step \(N\) is
  $$
  \mathcal{R}_N = \left\{ \sum_{i=0}^{N-1} A^{N-1-i} B\, u_i \right\} = \operatorname{range}(\, \mathcal{C}_N \,),
  $$
  where the \(N\)-step controllability matrix is
  $$
  \mathcal{C}_N = \begin{bmatrix} B & AB & \cdots & A^{N-1} B \end{bmatrix}.
  $$

- The discrete-time reachability Gramian is
  $$
  W_r(N) = \sum_{i=0}^{N-1} A^{i} B B^{\top} (A^{\top})^{i}.
  $$
  As in continuous time, \(x_N\) (from \(x_0=0\)) is reachable with finite energy iff \(x_N \in \operatorname{range}(W_r(N))\), and full reachability at step \(N\) holds iff \(W_r(N)\) is positive definite.

- One minimal-energy input sequence to reach \(x_N\) from \(x_0=0\):
  $$
  u_k^{\star} = B^{\top} (A^{\top})^{N-1-k} \, W_r(N)^{-1} \, x_N, \qquad k=0,\dots,N-1.
  $$
  The minimal energy is
  $$
  J^{\star} = \sum_{k=0}^{N-1} \lVert u_k^{\star} \rVert^2 = x_N^{\top} W_r(N)^{-1} x_N.
  $$

### Rank tests (controllability matrices)

- Continuous time: \((A,B)\) is controllable (hence reachable for any \(T>0\)) iff
  $$
  \operatorname{rank}\, \mathcal{C}_n = n, \qquad \mathcal{C}_n = \begin{bmatrix} B & AB & \cdots & A^{n-1} B \end{bmatrix}.
  $$
- Discrete time: same rank condition applies. If \(\operatorname{rank}(\mathcal{C}_n) = n\), then for any \(N \ge n\) every state is reachable in \(N\) steps.

### Example: double integrator (continuous time)

Let
$$
A = \begin{bmatrix} 0 & 1 \\ 0 & 0 \end{bmatrix},\quad B = \begin{bmatrix} 0 \\ 1 \end{bmatrix}.
$$
The controllability matrix \(\mathcal{C}_2 = \begin{bmatrix} 0 & 1 \\ 1 & 0 \end{bmatrix}\) has rank 2, so the system is controllable. The finite-horizon reachability Gramian is
$$
W_r(T) = \int_{0}^{T} e^{A\tau} B B^{\top} e^{A^{\top} \tau} d\tau =
\begin{bmatrix}
\tfrac{T^3}{3} & \tfrac{T^2}{2} \\
\tfrac{T^2}{2} & T
\end{bmatrix},\qquad \det W_r(T) = \tfrac{T^4}{12} > 0\ \ (T>0).
$$
Thus every state is reachable in any positive time, and the minimal energy to reach \(x_T\) from rest is \(x_T^{\top} W_r(T)^{-1} x_T\).

### Practical notes

- **Finite-horizon vs. global property**: Positive definiteness of \(W_r(T)\) for some \(T>0\) is equivalent to the rank test; both characterize controllability of \((A,B)\).
- **Numerics**: For stable \(A\), use continuous-time Lyapunov solvers for the infinite-horizon Gramian; for finite horizons, use quadrature or solve a differential Lyapunov equation backward in time.
- **Constraints**: Input/state constraints shrink the reachable set; the Gramian-based tests assume unconstrained inputs and linear dynamics.
