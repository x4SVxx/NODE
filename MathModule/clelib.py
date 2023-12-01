import numpy as np

# SYNCHRONIZATION ALGORITHM USING LINEAR KALMAN FILTER

def CS_filter(X, Dx, dt, t_M, t_S, R, T_max):

    Dn = 6.e-20
    Q  = 5.e-20/0.0225

    y = (t_S - t_M)%T_max - R

    # if abs(y) > 15:
    #     if y > 0:
    #         y -= T_max
    #     else:
    #         y += T_max

    F = np.array([[1., dt], [0., 1.]])
    G = np.array([[0.], [dt]])
    H = np.array([1., 0.])
    H = H.reshape(1, 2)
    x_ext = F.dot(X)
    D_ext = (F.dot(Dx)).dot(F.transpose()) + (G.dot(Q)).dot(G.transpose())
    K = (D_ext.dot(H.transpose()))/((H.dot(D_ext)).dot(H.transpose()) + Dn)
    Dx = D_ext - (K.dot(H)).dot(D_ext)
    X = x_ext + K.dot(y - H.dot(x_ext))
    nev = y - H.dot(x_ext)
    if abs(nev[0]) < 6*np.sqrt(Dx[0][0]):
        b = 1
    else:
        b = 0

    return b, X, Dx, nev


def make_initial(tx, rx, Range, T_max):
    t = []
    x = []
    for i in range(0, len(tx)):
        t.append(tx[i])
        if i > 0 and t[i] - t[i - 1] < 0:
            t[i] = t[i] + T_max
        x.append((rx[i] - tx[i])%T_max - Range)
    A = np.array([[len(tx), 0.], [0., 0.]])
    b = np.array([[0.0], [0.0]])
    for i in range(0, len(tx)):
        A[0][1] = A[0][1] + t[i]
        A[1][1] = A[1][1] + pow(t[i], 2)
        b[0][0] = b[0][0] + x[i]
        b[1][0] = b[1][0] + x[i] * t[i]
    A[1][0] = A[0][1]
    ax = (np.linalg.inv(A)).dot(b)
    delta = 0.
    for i in range(0, len(rx)):
        delta = delta + pow(ax[0][0] + ax[1][0] * t[i] - x[i], 2)
    delta = np.sqrt(delta / len(rx))
    if delta < 3.0e-10:
        flag = 1
    else:
        flag = 0
    return flag, ax[0][0], ax[1][0]


def check_PD(measurements, config):

    max_pd = 0.
    min_pd = 20.
    N = len(measurements)

    for i in range(N):
        if measurements[i]["data"]["corrected_timestamp"] > max_pd:
            max_pd = measurements[i]["data"]["corrected_timestamp"]
        if measurements[i]["data"]["corrected_timestamp"] < min_pd:
            min_pd = measurements[i]["data"]["corrected_timestamp"]

    if max_pd - min_pd > 5.:
        for i in range(1, N):
            if abs(measurements[i]["data"]["corrected_timestamp"] - measurements[0]["data"]["corrected_timestamp"]) > 10.:
                measurements[i]["data"]["corrected_timestamp"] -= np.sign(measurements[i]["data"]["corrected_timestamp"] - measurements[0]["data"]["corrected_timestamp"])*config.T_max
    return measurements


def coords_calc_2D(tag):
    N = len(tag.measurements)
    if N < 3:
        return False
    else:
        anc_pos = np.zeros((3, N))
        toa_m = np.zeros((N, 1))
        x0 = 0.
        y0 = 0.
        for i, msg in enumerate(tag.measurements):
            for anchor in tag.cle.anchors:
                if anchor.ID == msg["data"]["receiver"]:
                    anc_pos[0][i] = anchor.x
                    anc_pos[1][i] = anchor.y
                    anc_pos[2][i] = anchor.z
                    toa_m[i][0] = msg["data"]["corrected_timestamp"]
                    x0 += anc_pos[0][i] / N
                    y0 += anc_pos[1][i] / N
        toa_m = toa_m * tag.cfg.c
        Init = np.zeros((3, 1))
        Init[0, 0] = x0
        Init[1, 0] = y0
        Init[2, 0] = toa_m[0][0]
        try:
            b, X, DOP = solver_pd_2D(anc_pos, toa_m, tag.h, Init, tag.cfg)
            b = check_zone(anc_pos, toa_m, X, 3.)
            if b:
                tag.x = X[0, 0]
                tag.y = X[1, 0]
                tag.DOP = DOP
                return True
            else:
                return False
        except:
            return False


def solver_pd_2D(SatPos, PD, h, Init, config):
    N = PD.size
    y = PD
    X = Init
    k = 0
    while True:
        H = np.zeros((N, 3))
        Y = np.zeros((N, 1))
        for j in range(N):
            D = np.sqrt(pow(SatPos[0, j] - X[0, 0], 2) + pow(SatPos[1, j] - X[1, 0], 2) + pow(SatPos[2, j] - h, 2))
            H[j, 0] = (X[0, 0] - SatPos[0, j]) / D
            H[j, 1] = (X[1, 0] - SatPos[1, j]) / D
            H[j, 2] = 1.
            Y[j, 0] = D + X[2, 0]

        X_prev = X
        X = X + ((np.linalg.inv(H.transpose().dot(H))).dot(H.transpose())).dot(y-Y)
        k = k + 1

        if (np.linalg.norm(X - X_prev) < 0.001) or (k > 8):
            break
    invHH = np.linalg.inv(H.transpose().dot(H))
    DOP = np.sqrt(invHH[0, 0] * invHH[0, 0] + invHH[1, 1] * invHH[1, 1])
    if (np.linalg.norm(X - X_prev) < 1) and (np.sqrt(pow(X[0, 0], 2) + pow(X[1, 0], 2)) < config.zone):
        if np.linalg.norm(y - Y) < 1.:
            b = True
        else:
            b = False
    else:
        b = False

    return b, X, DOP


def medfilt1(input):
    output = []
    for i in range(1, len(input)-1):
        if input[i-1] <= input[i] and input[i-1] <= input[i+1]:
            if input[i] <= input[i+1]:
                output.append(input[i])
            else:
                output.append(input[i+1])

        if input[i] <= input[i-1] and input[i] <= input[i+1]:
            if input[i-1] <= input[i+1]:
                output.append(input[i-1])
            else:
                output.append(input[i+1])

        if input[i+1] <= input[i-1] and input[i+1] <= input[i]:
            if input[i-1] <= input[i]:
                output.append(input[i-1])
            else:
                output.append(input[i])
    return output

def check_zone(SatPos, pd, X, h):
    X_max = -1.e10
    X_min = 1.e10
    Y_max = -1.e10
    Y_min = 1.e10

    N = pd.size
    flag = True


    for j in range(N):
        if SatPos[0, j] < X_min:
            X_min = SatPos[0, j]
        if SatPos[0, j] > X_max:
            X_max = SatPos[0, j]
        if SatPos[1, j] < Y_min:
            Y_min = SatPos[1, j]
        if SatPos[1, j] > Y_max:
            Y_max = SatPos[1, j]

    # print("X_min = " + str(X_min))
    # print("X_max = " + str(X_max))
    # print("Y_min = " + str(Y_min))
    # print("Y_max = " + str(Y_max))

    if X[0] < X_min - h or X[0] > X_max + h:
        flag = False
    if X[1] < Y_min - h or X[1] > Y_max + h:
        flag = False

    # if not flag:
    #     print("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
    #     print(X)
    return flag
    