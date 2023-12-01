import clelib as cl
import time

class Tag():
    def __init__(self, msg, cle):
        self.organization = cle.organization
        self.roomid = cle.roomid
        self.cfg = cle.cfg
        self.cle = cle
        self.ID = msg['data']["sender"]
        self.SN = -1
        self.measurements = []
        self.x = None
        self.y = None
        self.x_f = None
        self.y_f = None
        self.h = cle.cfg.hei
        self.DOP = 1.
        self.starttime = time.time()
        self.lasttime = time.time()
        self.lifetime = 0.
        self.state = 0
        self.static = False
        self.SOS = False
        self.alpha = 0.5
        self.data2sendflag = 0
        self.anchors_number_to_solve = 0
        self.data2log = None
        self.x_buffer = []
        self.y_buffer = []
        self.buffer_length = cle.cfg.buffer_length
        self.accumulation_mode = cle.cfg.accumulation_mode

    def add_data(self, msg):
        TAG_STATIC = 0x1
        TAG_SOS = 0x8
        
        if 'data' in msg:
            if 'state' in msg['data']:
                if msg["data"]["state"]:
                    if msg["data"]["state"] & TAG_STATIC:
                        self.static = True
                    else:
                        self.static = False

                    if msg["data"]["state"] & TAG_SOS:
                        self.SOS = True
                    else:
                        self.SOS = False

            if time.time() - self.lasttime > 10:
                self.x_buffer = []
                self.y_buffer = []
                self.measurements = []
                self.SN = -1
                self.lasttime = time.time()
            if msg['data']["sn"] != self.SN:
                delta = msg['data']["sn"] - self.SN
                if delta < -240:
                    delta += 255
                if (delta > 0) or self.SN < 0:
                    self.SN = msg['data']["sn"]
                else:
                    return
                self.measurements = cl.check_PD(self.measurements, self.cfg)
                flag = cl.coords_calc_2D(self)
                if flag:
                    if self.accumulation_mode:
                        if time.time() - self.lasttime > 1:
                            self.data2sendflag = 1
                            self.x_buffer = []
                            self.y_buffer = []
                        else:
                            self.x_buffer.append(self.x)
                            self.y_buffer.append(self.y)
                            if len(self.x_buffer) == self.buffer_length:
                                sorted_x = sorted(self.x_buffer)
                                sorted_y = sorted(self.y_buffer)
                                self.x = sorted_x[round(self.buffer_length / 2)]
                                self.y = sorted_y[round(self.buffer_length / 2)]
                                self.data2sendflag = 1
                                self.x_buffer = []
                                self.y_buffer = []
                    else:
                        self.data2sendflag = 1

                    self.lasttime = time.time()
                    self.lifetime = self.lasttime - self.starttime

                    self.anchors_number_to_solve = len(self.measurements)

                else:
                    flag = 0
                if self.cfg.log:
                    self.log_cle_tag(flag)
                self.measurements = []
            self.measurements.append(msg)








